package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/chanzuckerberg/go-misc/lambda/tfe-metrics/state"
	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/go-tfe"
	"github.com/honeycombio/libhoney-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Runner struct {
	tfeToken          string
	honeycombDataset  string
	honeycombWriteKey string

	done bool

	stater state.Stater
}

func NewRunner(tfeToken, honeycombDataset, honeycombWriteKey string, stater state.Stater) *Runner {
	return &Runner{
		tfeToken:          tfeToken,
		honeycombDataset:  honeycombDataset,
		honeycombWriteKey: honeycombWriteKey,
		done:              false,

		stater: stater,
	}
}

// TODO move to go-misc version when merged
func SetupSentry(sentryDSN, env string) (func(), error) {
	if sentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:         sentryDSN,
			Environment: env,
		})
		if err != nil {
			f := func() {}
			return f, errors.Wrap(err, "Sentry initialization failed")
		}

		f := func() {
			sentry.Flush(time.Second * 5)
			sentry.Recover()
		}
		return f, nil
	}
	return func() {}, nil
}

func (r *Runner) RunOnce() error {
	logrus.Info("starting run")
	timeZero := time.Time{}
	config := tfe.DefaultConfig()
	config.Token = r.tfeToken

	client, err := tfe.NewClient(config)

	if err != nil {
		return errors.Wrap(err, "unable to create tfe client")
	}

	// set up honeycomb api
	err = libhoney.Init(libhoney.Config{
		WriteKey: r.honeycombWriteKey,
		Dataset:  r.honeycombDataset,
	})

	if err != nil {
		return errors.Wrap(err, "unable to create honeycomb client")
	}
	defer libhoney.Close()
	builder := libhoney.NewBuilder()

	for i := 0; i < 50; i++ {
		logrus.Infof("i %d", i)
		options := tfe.AdminRunsListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: i + 1,
			},
			RunStatus: tfe.String("planned_and_finished,applied,errored,discarded,policy_soft_failed,canceled,force_canceled"),
			Include:   "workspace,workspace.organization",
		}
		ctx := context.Background()
		list, err := client.AdminRuns.List(ctx, options)

		if err != nil {
			return errors.Wrap(err, "unable to list runs")
		}

		if list.Pagination.NextPage == 0 {
			break
		}

		for _, adminRun := range list.Items {
			isProcessed, err := r.stater.IsProcessed(adminRun.ID)
			logrus.Infof("run id: %s isprocessed: %t", adminRun.ID, isProcessed)
			if err != nil {
				return errors.Wrap(err, "err with state")
			}
			if !isProcessed {
				ev := builder.NewEvent()
				run, err := client.Runs.Read(ctx, adminRun.ID)
				if err != nil {
					return errors.Wrapf(err, "unable to read run from tf api %s", adminRun.ID)
				}

				event := map[string]interface{}{}

				ev.Timestamp = run.CreatedAt
				event["id"] = run.ID
				event["created_at"] = run.CreatedAt.Unix()
				event["has_changes"] = run.HasChanges
				event["is_destroy"] = run.IsDestroy
				event["message"] = run.Message
				event["source"] = run.Source
				event["status"] = run.Status

				if run.StatusTimestamps != nil {
					st := run.StatusTimestamps
					if run.CreatedAt != timeZero && st.PlanQueuabledAt != timeZero {
						event["pending_ms"] = run.CreatedAt.Sub(st.PlanQueuabledAt).Milliseconds()
					}
					if st.PlannedAt != timeZero && st.PlanningAt != timeZero {
						event["planning_ms"] = st.PlannedAt.Sub(st.PlanningAt).Milliseconds()
					}
					if st.AppliedAt != timeZero && st.ApplyingAt != timeZero {
						event["applying_ms"] = st.AppliedAt.Sub(st.ApplyingAt).Milliseconds()
					}
				}
				if run.Apply != nil {
					// TODO refactor go-tfe to allow including these resource on the Read call
					apply, _ := client.Applies.Read(ctx, run.Apply.ID)
					event["apply_id"] = run.Apply.ID
					// TODO   LogReadURL?
					event["apply_resource_additions"] = apply.ResourceAdditions
					event["apply_resource_changes"] = apply.ResourceChanges
					event["apply_resource_destructions"] = apply.ResourceDestructions
					event["apply_status"] = run.Status
					// TODO  StatusTimestamps
				}
				if run.ConfigurationVersion != nil {
					cv, _ := client.ConfigurationVersions.Read(ctx, run.ConfigurationVersion.ID)
					event["configuration_version_auto_queue_runs"] = cv.AutoQueueRuns
					event["configuration_version_error"] = cv.Error
					event["configuration_version_error_message"] = cv.ErrorMessage
					event["configuration_version_source"] = cv.Source
					event["configuration_version_speculative"] = cv.Speculative
					event["configuration_version_status"] = cv.Status
					// TODO cv.StatusTimestamps, cv.UploadURL
				}
				// TODO  CostEstimate
				if run.Plan != nil {
					plan, _ := client.Plans.Read(ctx, run.Plan.ID)
					event["plan_id"] = plan.ID
					event["plan_has_changes"] = plan.HasChanges
					// TODO   LogReadURL
					event["plan_resource_addtions"] = plan.ResourceAdditions
					event["plan_resource_changes"] = plan.ResourceChanges
					event["plan_resource_destructions"] = plan.ResourceDestructions
					event["plan_status"] = plan.Status
					// TODO   StatusTimestamps
					//  TODO  Exports
				}
				// TODO PolicyChecks
				if run.Workspace != nil {
					workspace, _ := client.Workspaces.ReadByID(ctx, run.Workspace.ID)
					event["workspace_id"] = run.Workspace.ID
					event["workspace_identifier"] = fmt.Sprintf("%s/%s", workspace.Organization.Name, workspace.Name)
					event["workspace_auto_apply"] = workspace.AutoApply
					event["workspace_environment"] = workspace.Environment
					event["workspace_file_triggers_enabled"] = workspace.FileTriggersEnabled
					event["workspace_migration_environment"] = workspace.MigrationEnvironment
					event["workspace_name"] = workspace.Name
					event["workspace_operations"] = workspace.Operations
					event["workspace_queue_all_runs"] = workspace.QueueAllRuns
					event["workspace_terraform_version"] = workspace.TerraformVersion
					event["repo_identifier"] = workspace.VCSRepo.Identifier
					event["workspace_working_directory"] = workspace.WorkingDirectory
					event["organization_name"] = workspace.Organization.Name
				}
				logrus.WithFields(event).Info("event")

				err = ev.Add(event)
				if err != nil {
					return errors.Wrap(err, "unable to add event to honeycomb")
				}

				err = ev.Send()
				if err != nil {
					return errors.Wrap(err, "unable to send event to honeycomb")
				}

				err = r.stater.SetProcessed(run.ID)
				if err != nil {
					return errors.Wrap(err, "unable to set processed")
				}
			}
		}
	}

	return nil
}
