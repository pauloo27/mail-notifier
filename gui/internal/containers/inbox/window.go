package inbox

import (
	"github.com/Pauloo27/mail-notifier/gui/utils"
	"github.com/Pauloo27/mail-notifier/socket/client"
	"github.com/Pauloo27/mail-notifier/socket/common/types"
	"github.com/gotk3/gotk3/gtk"
)

func Show(c *client.Client, box *types.Inbox, messages *types.CachedUnreadMessages) {
	gtk.Init(nil)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	utils.HandleError(err)

	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	mainContainer, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
	utils.HandleError(err)

	mainContainer.PackStart(createHeader(box, win), false, false, 5)
	mainContainer.PackStart(createMessageList(c, box, messages), true, true, 5)

	win.Add(mainContainer)

	win.SetTitle("Mail Notifier - Inbox")
	win.SetDefaultSize(450, 500)
	win.ShowAll()
	gtk.Main()
}
