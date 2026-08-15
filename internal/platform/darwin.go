//go:build darwin

package platform

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework QuartzCore
#import <Cocoa/Cocoa.h>
#import <QuartzCore/QuartzCore.h>

static void applyWindowConfig(double cornerRadius) {
	NSWindow *window = nil;
	for (NSWindow *w in [NSApp windows]) {
		if ([w isVisible]) {
			window = w;
			break;
		}
	}
	if (!window) window = [NSApp mainWindow];
	if (!window) window = [NSApp keyWindow];
	if (!window) return;

	// Hide traffic light buttons
	NSButton *close = [window standardWindowButton:NSWindowCloseButton];
	NSButton *mini  = [window standardWindowButton:NSWindowMiniaturizeButton];
	NSButton *zoom  = [window standardWindowButton:NSWindowZoomButton];
	if (close) close.hidden = YES;
	if (mini)  mini.hidden  = YES;
	if (zoom)  zoom.hidden  = YES;

	// Also hide their container view as fallback
	if (close && close.superview) {
		close.superview.hidden = YES;
	}

	[window setOpaque:NO];
	[window setBackgroundColor:[NSColor clearColor]];

	// Set corner radius on theme frame
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

	// Set corner radius on all subviews (NSVisualEffectView, etc.)
	for (NSView *subview in [contentView subviews]) {
		[subview setWantsLayer:YES];
		subview.layer.cornerRadius = cornerRadius;
		subview.layer.masksToBounds = YES;
	}
}

void configureWindowAppearance(double cornerRadius) {
	dispatch_after(
		dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.3 * NSEC_PER_SEC)),
		dispatch_get_main_queue(),
		^{ applyWindowConfig(cornerRadius); }
	);
}

void updateCornerRadius(double cornerRadius) {
	dispatch_async(dispatch_get_main_queue(), ^{
		applyWindowConfig(cornerRadius);
	});
}
*/
import "C"

func ConfigureWindow(cornerRadius float64) {
	C.configureWindowAppearance(C.double(cornerRadius))
}

func UpdateCornerRadius(cornerRadius float64) {
	C.updateCornerRadius(C.double(cornerRadius))
}
