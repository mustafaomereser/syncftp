//go:build cgo

// macOS penceresi: WKWebView, başlıksız görünüm (kapat/küçült düğmeleri sistemin), menü (Cmd+C/V için şart),
// dosya seçici ve indirme.
#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>
#include "window.h"

@interface SFHost : NSObject <NSApplicationDelegate, NSWindowDelegate, WKScriptMessageHandler, WKNavigationDelegate, WKUIDelegate, WKDownloadDelegate>
@property(strong) NSWindow *window;
@property(strong) WKWebView *web;
@property(copy) NSString *smoke;
@end

@implementation SFHost
- (BOOL)applicationShouldTerminateAfterLastWindowClosed:(NSApplication *)a { return YES; }

- (void)userContentController:(WKUserContentController *)c didReceiveScriptMessage:(WKScriptMessage *)m {
    if (![m.body isKindOfClass:[NSString class]]) return;
    NSString *s = m.body;
    if ([s isEqualToString:@"win:min"]) [self.window miniaturize:nil];
    else if ([s isEqualToString:@"win:max"]) [self.window zoom:nil];
    else if ([s isEqualToString:@"win:close"]) [self.window close];
    else if ([s isEqualToString:@"win:drag"]) { NSEvent *e = NSApp.currentEvent; if (e) [self.window performWindowDragWithEvent:e]; }
    else if ([s hasPrefix:@"open:"]) { NSURL *u = [NSURL URLWithString:[s substringFromIndex:5]]; if (u) [[NSWorkspace sharedWorkspace] openURL:u]; }
    else if ([s hasPrefix:@"smoke:"] && self.smoke) {
        [[s substringFromIndex:6] writeToFile:self.smoke atomically:YES encoding:NSUTF8StringEncoding error:nil];
        [NSApp terminate:nil];
    }
}

- (void)reportState {
    NSString *st = self.window.isZoomed ? @"max" : @"normal";
    [self.web evaluateJavaScript:[NSString stringWithFormat:@"window.__winState&&window.__winState('%@')", st] completionHandler:nil];
}
- (void)windowDidResize:(NSNotification *)n { [self reportState]; }

// <input type=file>
- (void)webView:(WKWebView *)w runOpenPanelWithParameters:(WKOpenPanelParameters *)p initiatedByFrame:(WKFrameInfo *)f completionHandler:(void (^)(NSArray<NSURL *> *))done {
    NSOpenPanel *o = [NSOpenPanel openPanel];
    o.allowsMultipleSelection = p.allowsMultipleSelection;
    o.canChooseFiles = YES;
    o.canChooseDirectories = p.allowsDirectories;
    [o beginSheetModalForWindow:self.window completionHandler:^(NSModalResponse r) { done(r == NSModalResponseOK ? o.URLs : nil); }];
}

// target=_blank: varsayılan tarayıcıda
- (WKWebView *)webView:(WKWebView *)w createWebViewWithConfiguration:(WKWebViewConfiguration *)c forNavigationAction:(WKNavigationAction *)a windowFeatures:(WKWindowFeatures *)x {
    if (a.request.URL) [[NSWorkspace sharedWorkspace] openURL:a.request.URL];
    return nil;
}

// Dosya indirme: ek olarak gelen ya da gösterilemeyen yanıt ~/Downloads'a
- (void)webView:(WKWebView *)w decidePolicyForNavigationResponse:(WKNavigationResponse *)r decisionHandler:(void (^)(WKNavigationResponsePolicy))d {
    BOOL attach = NO;
    if ([r.response isKindOfClass:[NSHTTPURLResponse class]]) {
        NSString *h = ((NSHTTPURLResponse *)r.response).allHeaderFields[@"Content-Disposition"];
        attach = h && [h.lowercaseString hasPrefix:@"attachment"];
    }
    if (@available(macOS 11.3, *)) {
        if (attach || !r.canShowMIMEType) { d(WKNavigationResponsePolicyDownload); return; }
    }
    d(WKNavigationResponsePolicyAllow);
}
- (void)webView:(WKWebView *)w navigationResponse:(WKNavigationResponse *)r didBecomeDownload:(WKDownload *)dl API_AVAILABLE(macos(11.3)) { dl.delegate = self; }
- (void)webView:(WKWebView *)w navigationAction:(WKNavigationAction *)a didBecomeDownload:(WKDownload *)dl API_AVAILABLE(macos(11.3)) { dl.delegate = self; }
- (void)download:(WKDownload *)dl decideDestinationUsingResponse:(NSURLResponse *)r suggestedFilename:(NSString *)name completionHandler:(void (^)(NSURL *))done API_AVAILABLE(macos(11.3)) {
    NSFileManager *fm = NSFileManager.defaultManager;
    NSURL *dir = [fm URLsForDirectory:NSDownloadsDirectory inDomains:NSUserDomainMask].firstObject;
    NSString *base = name.stringByDeletingPathExtension, *ext = name.pathExtension;
    NSURL *to = [dir URLByAppendingPathComponent:name];
    for (int i = 1; [fm fileExistsAtPath:to.path]; i++) {
        NSString *n = [NSString stringWithFormat:@"%@ (%d)", base, i];
        to = [dir URLByAppendingPathComponent:ext.length ? [n stringByAppendingPathExtension:ext] : n];
    }
    done(to);
}
@end

static void build_menu(NSString *name) {
    NSMenu *bar = [NSMenu new];
    NSMenuItem *appItem = [NSMenuItem new];
    [bar addItem:appItem];
    NSMenu *app = [NSMenu new];
    [app addItemWithTitle:[@"Hide " stringByAppendingString:name] action:@selector(hide:) keyEquivalent:@"h"];
    [app addItem:NSMenuItem.separatorItem];
    [app addItemWithTitle:[@"Quit " stringByAppendingString:name] action:@selector(terminate:) keyEquivalent:@"q"];
    appItem.submenu = app;

    NSMenuItem *editItem = [NSMenuItem new];
    [bar addItem:editItem];
    NSMenu *edit = [[NSMenu alloc] initWithTitle:@"Edit"];
    [edit addItemWithTitle:@"Undo" action:@selector(undo:) keyEquivalent:@"z"];
    [edit addItemWithTitle:@"Redo" action:@selector(redo:) keyEquivalent:@"Z"];
    [edit addItem:NSMenuItem.separatorItem];
    [edit addItemWithTitle:@"Cut" action:@selector(cut:) keyEquivalent:@"x"];
    [edit addItemWithTitle:@"Copy" action:@selector(copy:) keyEquivalent:@"c"];
    [edit addItemWithTitle:@"Paste" action:@selector(paste:) keyEquivalent:@"v"];
    editItem.submenu = edit;
    NSApp.mainMenu = bar;
}

int window_probe(void) { return 1; }

void window_run(const char *url, const char *title, const char *smoke) {
    @autoreleasepool {
        [NSApplication sharedApplication];
        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
        SFHost *host = [SFHost new];
        NSApp.delegate = host;
        NSString *t = [NSString stringWithUTF8String:title];
        build_menu(@"syncFTP");
        if (smoke && smoke[0]) host.smoke = [NSString stringWithUTF8String:smoke];

        NSString *js = @SYNCFTP_BRIDGE_JS("mac", "webkit.messageHandlers.syncftp.postMessage(m);");
        if (host.smoke) js = [js stringByAppendingString:@SYNCFTP_SMOKE_JS];
        WKWebViewConfiguration *cfg = [WKWebViewConfiguration new];
        [cfg.userContentController addScriptMessageHandler:host name:@"syncftp"];
        [cfg.userContentController addUserScript:[[WKUserScript alloc] initWithSource:js injectionTime:WKUserScriptInjectionTimeAtDocumentStart forMainFrameOnly:YES]];

        NSRect frame = NSMakeRect(0, 0, 1320, 860);
        NSWindow *win = [[NSWindow alloc] initWithContentRect:frame
            styleMask:NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | NSWindowStyleMaskMiniaturizable | NSWindowStyleMaskResizable | NSWindowStyleMaskFullSizeContentView
            backing:NSBackingStoreBuffered defer:NO];
        win.title = t;
        win.titlebarAppearsTransparent = YES;
        win.titleVisibility = NSWindowTitleHidden;
        win.backgroundColor = [NSColor colorWithSRGBRed:0.07 green:0.075 blue:0.09 alpha:1];
        win.minSize = NSMakeSize(900, 600);
        win.delegate = host;
        win.releasedWhenClosed = NO;
        [win center];
        host.window = win;

        WKWebView *web = [[WKWebView alloc] initWithFrame:frame configuration:cfg];
        web.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;
        web.navigationDelegate = host;
        web.UIDelegate = host;
        host.web = web;
        win.contentView = web;

        [web loadRequest:[NSURLRequest requestWithURL:[NSURL URLWithString:[NSString stringWithUTF8String:url]]]];
        [win makeKeyAndOrderFront:nil];
        [NSApp activateIgnoringOtherApps:YES];
        [NSApp run];
    }
}
