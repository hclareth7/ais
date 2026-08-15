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
		if (!window) return;

		// Hide native traffic light buttons
		[[window standardWindowButton:NSWindowCloseButton] setHidden:YES];
		[[window standardWindowButton:NSWindowMiniaturizeButton] setHidden:YES];
		[[window standardWindowButton:NSWindowZoomButton] setHidden:YES];

		[window setOpaque:NO];
		[window setBackgroundColor:[NSColor clearColor]];

		// Set corner radius on the theme frame (window chrome)
		NSView *contentView = [window contentView];
		NSView *themeFrame = [contentView superview];
		if (themeFrame) {
			[themeFrame setWantsLayer:YES];
			themeFrame.layer.cornerRadius = cornerRadius;
			themeFrame.layer.masksToBounds = YES;
		}

		// Set corner radius on content view
		[contentView setWantsLayer:YES];
		contentView.layer.cornerRadius = cornerRadius;
		contentView.layer.masksToBounds = YES;

		// Set corner radius on all subviews (NSVisualEffectView, WKWebView, etc.)
		for (NSView *subview in [contentView subviews]) {
			[subview setWantsLayer:YES];
			subview.layer.cornerRadius = cornerRadius;
			subview.layer.masksToBounds = YES;
		}
	});
}
*/
import "C"

func ConfigureWindow(cornerRadius float64) {
	C.configureWindowAppearance(C.double(cornerRadius))
}
