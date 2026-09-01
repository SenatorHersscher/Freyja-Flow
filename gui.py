#!/usr/bin/env python3
import sys
import os
import signal

# Process ve Uygulama Adını 'freyja-flow' olarak mühürle (KDE'de 'python3' yazmasını engeller)
sys.argv[0] = "freyja-flow"

import gi
gi.require_version('Gtk', '3.0')
gi.require_version('WebKit2', '4.1')
from gi.repository import Gtk, WebKit2, GdkPixbuf, GLib

GLib.set_prgname("freyja-flow")
GLib.set_application_name("Freyja Flow")

class FreyjaFlowWindow(Gtk.Window):
    def __init__(self, url):
        super().__init__(title="Freyja Flow")
        self.set_wmclass("freyja-flow", "Freyja Flow")
        self.set_role("FreyjaFlow")
        self.set_default_size(1280, 820)
        self.set_position(Gtk.WindowPosition.CENTER)
        
        # Pencere ikonu
        icon_path = os.path.join(os.path.dirname(__file__), "frontend/assets/icon.png")
        if os.path.exists(icon_path):
            try:
                pixbuf = GdkPixbuf.Pixbuf.new_from_file_at_scale(icon_path, 64, 64, True)
                self.set_icon(pixbuf)
            except Exception:
                pass

        # WebKit2 ayarları
        settings = WebKit2.Settings()
        settings.set_enable_developer_extras(False)
        settings.set_enable_webgl(True)
        settings.set_enable_accelerated_2d_canvas(True)
        
        self.webview = WebKit2.WebView.new_with_settings(settings)
        self.webview.load_uri(url)
        self.add(self.webview)
        
        self.connect("destroy", Gtk.main_quit)
        self.show_all()

if __name__ == "__main__":
    signal.signal(signal.SIGINT, signal.SIG_DFL)
    target_url = sys.argv[1] if len(sys.argv) > 1 else "http://localhost:8090"
    win = FreyjaFlowWindow(target_url)
    Gtk.main()
