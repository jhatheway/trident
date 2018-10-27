package TriUI

import (
	"encoding/json"
	"html/template"
	"path/filepath"
	"strconv"
	"time"
	tr "lib"
)

func WikiUI_ApplyModOpts(cui PfUI, wiki *tr.PfWikiPage) {
	opts := tr.Wiki_GetModOpts(cui)
	op := wiki.FullPath
	np := tr.URL_Append(opts.URLroot, op[len(opts.Pathroot):])
	np = tr.URL_Append(opts.URLpfx, np)
	wiki.FullPath = np

	isdir := wiki.Path[len(wiki.Path)-1] == '/'

	wiki.Path = filepath.Base(wiki.Path)

	if isdir {
		wiki.Path += "/"
	}
}

func WikiUI_ApplyModOptsMulti(cui PfUI, wikis []tr.PfWikiPage) {
	for i := range wikis {
		WikiUI_ApplyModOpts(cui, &wikis[i])
	}
}

func h_wiki_edit(cui PfUI) {
	path := cui.GetSubPath()
	rev := cui.GetArg("rev")

	var m tr.PfWikiMarkdown
	err := m.Fetch(cui, path, rev)
	if err != nil && err != ErrNoRows {
		H_error(cui, StatusBadRequest)
		return
	}

	/* Output the page */
	type Page struct {
		*PfPage
		WikiText string
	}

	p := Page{cui.Page_def(), m.Markdown}
	p.AddJS("misc")
	p.AddJS("editor")
	cui.Page_show("wiki/edit.tmpl", p)
}

func h_wiki_source(cui PfUI) {
	var m tr.PfWikiMarkdown
	var h tr.PfWikiHTML
	var err error

	path := cui.GetSubPath()
	rev := cui.GetArg("rev")

	err = m.Fetch(cui, path, rev)
	if err != nil && err != ErrNoRows {
		H_error(cui, StatusBadRequest)
		return
	}

	err = h.Fetch(cui, path, "")
	if err != nil && err != ErrNoRows {
		H_error(cui, StatusBadRequest)
		return
	}

	/* Output the page */
	type Page struct {
		*PfPage
		WikiText string
		WikiHTML template.HTML
	}

	p := Page{cui.Page_def(), m.Markdown, h.HTML_Body}
	cui.Page_show("wiki/source.tmpl", p)
}

func h_wiki_raw(cui PfUI) {
	var m tr.PfWikiMarkdown
	var err error

	path := cui.GetSubPath()
	rev := cui.GetArg("rev")

	err = m.Fetch(cui, path, rev)
	if err != nil && err != ErrNoRows {
		H_error(cui, StatusBadRequest)
		return
	}

	fname := tr.Wiki_Title(path) + ".md"

	/* Output the page (RFC7763 for the MIMEType) */
	cui.SetContentType("text/markdown")
	cui.SetFileName(fname)
	cui.SetExpires(60)
	cui.SetRaw([]byte(m.Markdown))
}

func h_wiki_diff(cui PfUI) {
	path := cui.GetSubPath()
	revA := cui.GetArg("rev")
	revB := cui.GetArg("revB")
	if revA == "" || revB == "" {
		H_error(cui, StatusBadRequest)
		return
	}

	diff, err := tr.Wiki_Diff(cui, path, revA, revB)
	if err != nil {
		cui.Errf("Wiki_Diff(%q,%q,%q): %s", path, revA, revB, err.Error())
		H_NoAccess(cui)
		return
	}

	/* Output the page */
	type Page struct {
		*PfPage
		RevA string
		RevB string
		Diff []tr.PfDiff
	}

	p := Page{cui.Page_def(), revA, revB, diff}
	cui.Page_show("wiki/diff.tmpl", p)
}

func h_wiki_read(cui PfUI) {
	path := cui.GetSubPath()
	rev := cui.GetArg("rev")

	var h tr.PfWikiHTML
	err := h.Fetch(cui, path, rev)
	if err != nil && err != ErrNoRows {
		H_error(cui, StatusBadRequest)
		return
	}

	/* Output the page */
	type Page struct {
		*PfPage
		CanEdit  bool
		WikiTOC  template.HTML
		WikiHTML template.HTML
		LastWhen time.Time
		LastUser string
		LastName string
	}

	p := Page{cui.Page_def(), true, h.HTML_TOC, h.HTML_Body, h.Entered, h.UserName, h.FullName}
	cui.Page_show("wiki/read.tmpl", p)
}

func h_wiki_history(cui PfUI) {
	var err error
	var revs []tr.PfWikiRev

	path := cui.GetSubPath()
	pageSize := tr.PAGER_PERPAGE /* TODO: Eventually I'd like this to come in from a parameter */

	total := 0
	offset := 0

	offset_v, err := cui.FormValue("offset")
	if err == nil && offset_v != "" {
		offset, _ = strconv.Atoi(offset_v)
	}

	total, err = tr.Wiki_RevisionMax(cui, path)
	if err != nil {
		H_error(cui, StatusBadRequest)
		return
	}

	revs, err = tr.Wiki_RevisionList(cui, path, offset, total)
	if err != nil {
		H_error(cui, StatusBadRequest)
		return
	}

	/* Output the page */
	type Page struct {
		*PfPage
		PageSize    int
		LastPage    int
		PagerOffset int
		PagerTotal  int
		Search      string
		Revs        []tr.PfWikiRev
	}

	p := Page{cui.Page_def(), pageSize, tr.Template_Pager_LastPage(total, pageSize), offset, total, "", revs}
	cui.Page_show("wiki/history.tmpl", p)
}

func h_wiki_search(cui PfUI) {
	var res []tr.PfWikiResult

	total := 0
	offset := 0
	pageSize := tr.PAGER_PERPAGE /* TODO: Eventually I'd like this to come in from a parameter */

	search, err := cui.FormValue("q")

	if err == nil && search != "" {
		offset_v, err2 := cui.FormValue("offset")
		if err2 == nil && offset_v != "" {
			offset, _ = strconv.Atoi(offset_v)
		}

		total, err = tr.Wiki_SearchMax(cui, search)
		if err != nil {
			H_error(cui, StatusBadRequest)
			return
		}

		res, err = tr.Wiki_SearchList(cui, search, offset, total)
		if err != nil {
			H_error(cui, StatusBadRequest)
			return
		}
	}

	/* Output the page */
	type popt struct {
		Q      string `label:"Search Query" hint:"What you are looking for" htmlclass:"search"`
		Button string `label:"Search" hint:"To look for things" pftype:"submit" pfcol:""`
	}

	type Page struct {
		*PfPage
		Search      popt
		PageSize    int
		LastPage    int
		PagerOffset int
		PagerTotal  int
		PathPrefix  string
		Results     []tr.PfWikiResult
	}

	mopts := tr.Wiki_GetModOpts(cui)
	opt := popt{search, ""}
	p := Page{cui.Page_def(), opt, pageSize, tr.Template_Pager_LastPage(total, pageSize), offset, total, mopts.URLroot, res}
	cui.Page_show("wiki/search.tmpl", p)
}

func h_wiki_children(cui PfUI) {
	var wikis []tr.PfWikiPage

	path := cui.GetSubPath()
	pageSize := tr.PAGER_PERPAGE /* TODO: Eventually I'd like this to come in from a parameter */

	total := 0
	offset := 0

	offset_v, err := cui.FormValue("offset")
	if err == nil && offset_v != "" {
		offset, _ = strconv.Atoi(offset_v)
	}

	total, err = tr.Wiki_ChildPagesMax(cui, path)
	if err != nil {
		H_error(cui, StatusBadRequest)
		return
	}

	wikis, err = tr.Wiki_ChildPagesList(cui, path, offset, total)
	if err != nil {
		H_error(cui, StatusBadRequest)
		return
	}

	WikiUI_ApplyModOptsMulti(cui, wikis)

	/* Output the page */
	type Page struct {
		*PfPage
		PageSize    int
		LastPage    int
		PagerOffset int
		PagerTotal  int
		Search      string
		Paths       []tr.PfWikiPage
	}

	p := Page{cui.Page_def(), pageSize, tr.Template_Pager_LastPage(total, pageSize), offset, total, "", wikis}
	cui.Page_show("wiki/children.tmpl", p)
}

func h_wiki_options(cui PfUI) {
	var err error

	path := cui.GetSubPath()

	type move struct {
		Path     string `label:"New path of the page" pfreq:"yes"`
		Children bool   `label:"Move all children of this page?" hint:"Only applies when the page has children"`
		Confirm  bool   `label:"Confirm Moving" pfreq:"yes"`
		Button   string `label:"Move Page" pftype:"submit"`
		Message  string /* Used by pfform() */
		Error    string /* Used by pfform() */
	}

	type del struct {
		Children bool   `label:"Delete all children of this page?" hint:"Only applies when the page has children"`
		Confirm  bool   `label:"Confirm Deletion" pfreq:"yes"`
		Button   string `label:"Delete Page" pftype:"submit" htmlclass:"deny"`
		Message  string /* Used by pfform() */
		Error    string /* Used by pfform() */
	}

	type cpy struct {
		Path     string `label:"Path of the page" pfreq:"yes"`
		Children bool   `label:"Copy all children of this page?" hint:"Only applies when the page has children"`
		Confirm  bool   `label:"Confirm copying" pfreq:"yes"`
		Button   string `label:"Copy Page" pftype:"submit"`
		Message  string /* Used by pfform() */
		Error    string /* Used by pfform() */
		cui      PfUI   /* Used by pfform() which calls FormContext() */
	}

	m := move{path, true, false, "", "", ""}
	d := del{true, false, "", "", ""}
	c := cpy{path, true, false, "", "", "", cui}

	if cui.IsPOST() {
		button, err1 := cui.FormValue("button")
		confirmed, err2 := cui.FormValue("confirm")
		children, err3 := cui.FormValue("children")
		newpath, err4 := cui.FormValue("path")

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			m.Error = "Invalid input"
			button = "Invalid"
		}

		if children == "on" {
			children = "yes"
		} else {
			children = "no"
		}

		switch button {
		case "Move Page":
			if confirmed != "on" {
				m.Error = "Did not confirm"
			} else {
				mopts := tr.Wiki_GetModOpts(cui)
				cmd := mopts.Cmdpfx + " move"
				arg := []string{path, newpath, children}

				_, err = cui.HandleCmd(cmd, arg)

				if err != nil {
					m.Error = err.Error()
				} else {
					url := tr.URL_Append(mopts.URLroot, newpath)
					cui.SetRedirect(url, StatusSeeOther)
					return
				}
			}
			break

		case "Delete Page":
			if confirmed != "on" {
				d.Error = "Did not confirm"
			} else {
				mopts := tr.Wiki_GetModOpts(cui)
				cmd := mopts.Cmdpfx + " delete"
				arg := []string{path, children}

				_, err = cui.HandleCmd(cmd, arg)

				if err != nil {
					d.Error = err.Error()
				} else {
					url := "../"
					cui.SetRedirect(url, StatusSeeOther)
					return
				}
			}
			break

		case "Copy Page":
			if confirmed != "on" {
				d.Error = "Did not confirm"
			} else {
				mopts := tr.Wiki_GetModOpts(cui)
				cmd := mopts.Cmdpfx + " copy"
				arg := []string{path, newpath, children}

				_, err = cui.HandleCmd(cmd, arg)

				if err != nil {
					c.Error = err.Error()
				} else {
					c.Message = "Page copied"
					return
				}
			}
			break
		}
	}

	type Page struct {
		*PfPage
		Move   move
		Delete del
		Copy   cpy
	}

	p := Page{cui.Page_def(), m, d, c}
	cui.Page_show("wiki/options.tmpl", p)
}

func h_wiki_newpage(cui PfUI) {
	path := cui.GetSubPath()

	l := len(path)
	if l > 0 && path[l-1] != '/' {
		path += "/"
	}

	if cui.IsPOST() {
		curpath, err := cui.FormValue("curpath")
		page, err2 := cui.FormValue("page")
		if err == nil && err2 == nil {
			mopts := tr.Wiki_GetModOpts(cui)
			url := tr.URL_Append(mopts.URLroot, curpath)
			url = tr.URL_Append(url, page)
			url += "?s=edit"
			cui.SetRedirect(url, StatusSeeOther)
			return
		}
	}

	type np struct {
		CurPath string `label:"Current Path of the page" pfset:"nobody" pfget:"user"`
		Page    string `label:"Name of new page" pfreq:"yes" hint:"Can include '/' to create multiple sub-levels"`
		Button  string `label:"Create New Page" pftype:"submit"`
	}

	type Page struct {
		*PfPage
		Opt np
	}

	p := Page{cui.Page_def(), np{path, "", ""}}
	cui.Page_show("wiki/newpage.tmpl", p)
}

func wiki_post_ajax(cui PfUI, path string) {
	rawbody := cui.GetBody()

	type Wiki struct {
		Markdown string `json:"markdown"`
		Message  string `json:"message"`
	}

	var body Wiki

	err := json.Unmarshal(rawbody, &body)
	if err != nil {
		cui.JSONAnswer("error", "JSON parsing failed")
		return
	}

	title := tr.Wiki_Title(path)
	mopts := tr.Wiki_GetModOpts(cui)
	cmd := mopts.Cmdpfx + " update"
	arg := []string{path, body.Message, title, body.Markdown}

	_, err = cui.CmdOut(cmd, arg)
	if err != nil {
		cui.JSONAnswer("error", "Update failed: "+err.Error())
		return
	}

	cui.JSONAnswer("ok", "Updated")
}

func wiki_post_form(cui PfUI, path string) (err error) {
	title := tr.Wiki_Title(path)

	mopts := tr.Wiki_GetModOpts(cui)
	cmd := mopts.Cmdpfx + " update"
	arg := []string{path, "", title, ""}

	_, err = cui.HandleCmd(cmd, arg)
	return
}

func H_wiki(cui PfUI) {
	var err error

	/* URL of the page */
	cui.SetSubPath("/" + cui.GetPathString())

	for _, p := range cui.GetPath() {
		cui.AddCrumb(p, p, "")
	}

	sub := cui.GetArg("s")

	/* Ajax Post? */
	if sub == "post" {
		wiki_post_ajax(cui, cui.GetSubPath())
		return
	}

	if sub == "edit" && cui.IsPOST() {
		err = wiki_post_form(cui, cui.GetSubPath())
		if err == nil {
			cui.SetRedirect("?p=read", StatusSeeOther)
			return
		}
	}

	var menu = NewPfUIMenu([]PfUIMentry{
		{"", "", PERM_USER, h_wiki_read, nil},
		{"?s=read", "Read", PERM_USER, h_wiki_read, nil},
		{"?s=source", "Source", PERM_USER, h_wiki_source, nil},
		{"?s=raw", "Raw", PERM_USER + PERM_HIDDEN, h_wiki_raw, nil},
		{"?s=edit", "Edit", PERM_USER, h_wiki_edit, nil},
		{"?s=history", "History", PERM_USER, h_wiki_history, nil},
		{"?s=diff", "Diff", PERM_USER | PERM_HIDDEN, h_wiki_diff, nil},
		{"?s=options", "Options", PERM_USER, h_wiki_options, nil},
		{"?s=children", "Child Pages", PERM_USER, h_wiki_children, nil},
		{"?s=newpage", "New Page", PERM_USER, h_wiki_newpage, nil},
		{"?s=search", "Search", PERM_USER, h_wiki_search, nil},
	})

	if sub == "read" {
		sub = ""
	}

	if sub != "" {
		sub = "?s=" + sub
	}

	cui.MenuPath(menu, &[]string{sub})
}
