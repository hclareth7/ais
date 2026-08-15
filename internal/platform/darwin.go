//go:build darwin

package platform

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework QuartzCore
#import <Cocoa/Cocoa.h>
#import <QuartzCore/QuartzCore.h>

void configureWindowAppearance(double cornerRadius) {
	dispatch_async(dispatch_get_main_queue(), ^{
		NSWindow *window = [NSApp mainWindow];
		if (!window) {
			window = [NSApp keyWindow];
		}
		if (window) {
			[window setOpaque:NO];
			[window setBackgroundColor:[NSColor clearColor]];
			NSView *contentView = [window contentView];
			[contentView setWantsLayer:YES];
			contentView.layer.cornerRadius = cornerRadius;
			contentView.layer.masksToBounds = YES;
		}
	});
}
*/
import "C"

func ConfigureWindow(cornerRadius float64) {
	C.configureWindowAppearance(C.double(cornerRadius))
}
