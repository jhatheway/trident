package TriUI

import (
	"html/template"
	"net/http"
	tr "lib"
)

type PfRootUI interface {
	H_root(w http.ResponseWriter, r *http.Request)
}

type PfRootUIS struct {
	newui PfNewUI
	PfRootUI
}

func NewPfRootUI(newui PfNewUI) (o PfRootUI) {
	return &PfRootUIS{newui: newui}
}

func (o *PfRootUIS) New() PfUI {
	return o.newui()
}

func h_index(cui PfUI) {
	type Page struct {
		*PfPage
		About template.HTML
	}

	/* TODO: Render Welcome using Markdown renderer */
	val := tr.System_Get().Welcome
	safe := NewlineBR(val)

	p := Page{cui.Page_def(), safe}
	p.HeaderImg = tr.System_Get().HeaderImg
	cui.Page_show("index.tmpl", p)
}

func h_robots(cui PfUI) {
	if tr.System_Get().NoIndex {
		h_static_file(cui, "robots.txt")
	} else {
		h_static_file(cui, "robots-ok.txt")
	}
}

/* Root page -- where Go's net/http gives it to us */
func (o *PfRootUIS) H_root(w http.ResponseWriter, r *http.Request) {
	cui := o.New()

	err := cui.UIInit(w, r)
	if err != nil {
		cui.Err(err.Error())
		H_error(cui, StatusBadRequest)
		cui.Flush()
		return
	}

	/* Set cancelation signal */
	abort := w.(http.CloseNotifier).CloseNotify()
	cui.SetAbort(abort)

	path := cui.GetPath()

	/* Get the Client IP & remote address */
	err = cui.SetClientIP()
	if err != nil {
		/* Something wrong with figuring out who they are */
		cui.Errf("SetClientIP: %s", err.Error())
		H_error(cui, StatusServiceUnavailable)
		cui.Flush()
		return
	}

	/* Check for static files/dirs */
	statics := []string{"favicon.ico", "css", "gfx", "js"}
	for _, p := range statics {
		if path[0] == p {
			h_static(cui)
			cui.Flush()
			return
		}
	}

	/*
	 * Homedirectory redirect:
	 * https://example.net/~username/ redirects to /user/username/
	 */
	if len(path[0]) > 0 && path[0][0] == '~' {
		cui.SetRedirect("/user/"+path[0][1:]+"/", StatusFound)
		cui.Flush()
		return
	}

	/* Initialize the token */
	cui.InitToken()

	/* The main menu */
	menu := NewPfUIMenu([]PfUIMentry{
		{"", "Home", PERM_NONE, h_index, nil},

		/* Service Discovery */
		{".well-known", "", PERM_NONE, h_wellknown, nil},

		/* Choice files */
		{"robots.txt", "", PERM_NONE | PERM_NOSUBS, h_robots, nil},

		/* QR Codes */
		{"qr", "", PERM_USER, h_qr, nil},

		/* From mainmenu (hidden as they are shown there) */
		{"user", "User", PERM_USER | PERM_HIDDEN, h_user, nil},
		{"group", "Group", PERM_USER | PERM_HIDDEN, h_group, nil},
		{"system", "System", PERM_SYS_ADMIN | PERM_HIDDEN, h_system, nil},

		/* Extras */
		{"search", "Search", PERM_USER | PERM_HIDDEN, h_search, nil},
		{"cli", "CLI", PERM_CLI, h_cli, nil},
		{"api", "", PERM_LOOPBACK | PERM_API, h_api, nil},
		{"oauth2", "OAuth2", PERM_USER, h_oauth, nil},
		{"login", "Login", PERM_NONE | PERM_USER | PERM_NOSUBS, h_login, nil},
		{"logout", "Logout", PERM_NONE | PERM_USER | PERM_HIDDEN | PERM_NOSUBS, h_logout, nil},
		{"recover", "Password Recover", PERM_NONE | PERM_HIDDEN, h_recover, nil},
	})

	
	/* JAMES TODO I think the below line can be deleted */
	// menu.Add(PfUIMentry)
	cui.UIMenu(menu)

	/* Flush it all to the client */
	cui.Flush()
}

func h_root(cui PfUI, menu *PfUIMenu) {
	menu.Add(PfUIMentry{"recover", "Password Recover", PERM_NONE | PERM_HIDDEN, h_recover, nil})
}

