//go:build cgo

// Linux penceresi: GTK3 + WebKit2GTK 4.1, çerçevesiz; taşıma ve boyutlandırma sayfadan istenir (titlebar.js).
#include <gtk/gtk.h>
#include <webkit2/webkit2.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include "window.h"

static GtkWindow *win;
static WebKitWebView *web;
static const char *smoke_path;
static guint press_button = 1;
static gint press_x, press_y;
static guint32 press_time = GDK_CURRENT_TIME;

static gboolean on_press(GtkWidget *w, GdkEventButton *e, gpointer d) {
    press_button = e->button;
    press_x = (gint)e->x_root;
    press_y = (gint)e->y_root;
    press_time = e->time;
    return FALSE;
}

static gboolean on_state(GtkWidget *w, GdkEventWindowState *e, gpointer d) {
    gboolean max = (e->new_window_state & GDK_WINDOW_STATE_MAXIMIZED) != 0;
    char js[96];
    snprintf(js, sizeof js, "window.__winState&&window.__winState('%s')", max ? "max" : "normal");
    webkit_web_view_evaluate_javascript(web, js, -1, NULL, NULL, NULL, NULL, NULL);
    return FALSE;
}

static GdkWindowEdge edge_of(const char *s) {
    if (!strcmp(s, "nw")) return GDK_WINDOW_EDGE_NORTH_WEST;
    if (!strcmp(s, "ne")) return GDK_WINDOW_EDGE_NORTH_EAST;
    if (!strcmp(s, "sw")) return GDK_WINDOW_EDGE_SOUTH_WEST;
    if (!strcmp(s, "se")) return GDK_WINDOW_EDGE_SOUTH_EAST;
    if (!strcmp(s, "n")) return GDK_WINDOW_EDGE_NORTH;
    if (!strcmp(s, "s")) return GDK_WINDOW_EDGE_SOUTH;
    if (!strcmp(s, "w")) return GDK_WINDOW_EDGE_WEST;
    return GDK_WINDOW_EDGE_EAST;
}

static void on_message(WebKitUserContentManager *m, WebKitJavascriptResult *r, gpointer d) {
    char *s = jsc_value_to_string(webkit_javascript_result_get_js_value(r));
    if (!strcmp(s, "win:min")) gtk_window_iconify(win);
    else if (!strcmp(s, "win:max")) { if (gtk_window_is_maximized(win)) gtk_window_unmaximize(win); else gtk_window_maximize(win); }
    else if (!strcmp(s, "win:close")) gtk_window_close(win);
    else if (!strcmp(s, "win:drag")) gtk_window_begin_move_drag(win, press_button, press_x, press_y, press_time);
    else if (!strncmp(s, "win:resize:", 11)) gtk_window_begin_resize_drag(win, edge_of(s + 11), press_button, press_x, press_y, press_time);
    else if (!strncmp(s, "open:", 5)) g_app_info_launch_default_for_uri(s + 5, NULL, NULL);
    else if (!strncmp(s, "smoke:", 6) && smoke_path) {
        FILE *f = fopen(smoke_path, "w");
        if (f) { fputs(s + 6, f); fclose(f); }
        gtk_main_quit();
    }
    g_free(s);
}

// target=_blank: varsayılan tarayıcıda
static GtkWidget *on_create(WebKitWebView *v, WebKitNavigationAction *a, gpointer d) {
    const char *uri = webkit_uri_request_get_uri(webkit_navigation_action_get_request(a));
    if (uri) g_app_info_launch_default_for_uri(uri, NULL, NULL);
    return NULL;
}

int window_probe(void) { return gtk_init_check(NULL, NULL) ? 1 : 0; }

void window_run(const char *url, const char *title, const char *smoke) {
    gtk_init(NULL, NULL);
    // Bazı sürücülerde (NVIDIA) WebKit'in DMA-BUF yolu boş ekran veriyor
    setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1", 0);
    smoke_path = (smoke && smoke[0]) ? smoke : NULL;

    win = GTK_WINDOW(gtk_window_new(GTK_WINDOW_TOPLEVEL));
    gtk_window_set_title(win, title);
    gtk_window_set_default_size(win, 1320, 860);
    gtk_window_set_decorated(win, FALSE);
    gtk_window_set_position(win, GTK_WIN_POS_CENTER);
    gtk_window_set_icon_name(win, "syncftp");
    GdkGeometry g = {.min_width = 900, .min_height = 600};
    gtk_window_set_geometry_hints(win, NULL, &g, GDK_HINT_MIN_SIZE);

    WebKitUserContentManager *ucm = webkit_user_content_manager_new();
    g_signal_connect(ucm, "script-message-received::syncftp", G_CALLBACK(on_message), NULL);
    webkit_user_content_manager_register_script_message_handler(ucm, "syncftp");
    GString *js = g_string_new(SYNCFTP_BRIDGE_JS("gtk", "webkit.messageHandlers.syncftp.postMessage(m);"));
    if (smoke_path) g_string_append(js, SYNCFTP_SMOKE_JS);
    webkit_user_content_manager_add_script(ucm, webkit_user_script_new(js->str, WEBKIT_USER_CONTENT_INJECT_TOP_FRAME, WEBKIT_USER_SCRIPT_INJECT_AT_DOCUMENT_START, NULL, NULL));
    g_string_free(js, TRUE);

    web = WEBKIT_WEB_VIEW(webkit_web_view_new_with_user_content_manager(ucm));
    WebKitSettings *st = webkit_web_view_get_settings(web);
    webkit_settings_set_enable_developer_extras(st, FALSE);
    g_signal_connect(web, "create", G_CALLBACK(on_create), NULL);
    g_signal_connect(web, "button-press-event", G_CALLBACK(on_press), NULL);
    g_signal_connect(win, "window-state-event", G_CALLBACK(on_state), NULL);
    g_signal_connect(win, "destroy", G_CALLBACK(gtk_main_quit), NULL);

    gtk_container_add(GTK_CONTAINER(win), GTK_WIDGET(web));
    webkit_web_view_load_uri(web, url);
    gtk_widget_show_all(GTK_WIDGET(win));
    gtk_main();
}
