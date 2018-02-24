#ifndef SHARENIX_CLIPBOARD_H
#define SHARENIX_CLIPBOARD_H

#include <gtk/gtk.h>

/* terrible hack to add set_can_store
   TODO: PR proper implementation to go-gtk */

static void _gtk_clipboard_set_can_store(void* clip) {
    gtk_clipboard_set_can_store((GtkClipboard*)clip, 0, 0);
}

#endif
