package osutil

import (
	"os"
	"runtime"
)

func IsDesktopEnvironment() bool {
    switch runtime.GOOS {
    case "darwin":
        // macOS always has a desktop environment if it's running
        return true
    
    case "linux":
        // Check for GNOME specifically
        if os.Getenv("XDG_CURRENT_DESKTOP") == "GNOME" {
            return true
        }
        
        // More permissive: check for any desktop environment
        if os.Getenv("XDG_CURRENT_DESKTOP") != "" {
            return true
        }
        
        // Check for X11 or Wayland display servers
        if os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != "" {
            return true
        }
        
        return false
    
    default:
        return false
    }
}