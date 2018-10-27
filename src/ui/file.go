package TriUI

import (
	"errors"
	"strconv"
	tr "lib"
)

func h_file_history(cui PfUI) {
	var err error
	var revs []tr.PfFile

	path := cui.GetSubPath()

	total := 0
	offset := 0
	pageSize := tr.PAGER_PERPAGE /* TODO: Eventually I'd like this to come in from a parameter */

	offset_v, err := cui.FormValue("offset")
	if err == nil && offset_v != "" {
		offset, _ = strconv.Atoi(offset_v)
	}

	total, err = tr.File_RevisionMax(cui, path)
	if err != nil {
		H_error(cui, 500)
		return
	}

	revs, err = tr.File_RevisionList(cui, path, offset, total)
	if err != nil {
		H_error(cui, 500)
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
		Revs        []tr.PfFile
	}

	p := Page{cui.Page_def(), pageSize, tr.Template_Pager_LastPage(total, pageSize), offset, total, "", revs}
	cui.Page_show("file/history.tmpl", p)
}

func FileUIApplyModOpts(cui PfUI, file *tr.PfFile) {
	opts := tr.File_GetModOpts(cui)
	op := file.FullPath
	np := tr.URL_Append(opts.URLroot, op[len(opts.Pathroot):])
	np = tr.URL_Append(opts.URLpfx, np)
	file.FullPath = np
}

func FileUIApplyModOptsM(cui PfUI, files []tr.PfFile) {
	for i := range files {
		FileUIApplyModOpts(cui, &files[i])
	}
}

func h_file_list_dir(cui PfUI) {
	path := cui.GetSubPath()

	total := 0
	offset := 0
	pageSize := tr.PAGER_PERPAGE /* TODO: Eventually I'd like this to come in from a parameter */

	offset_v, err := cui.FormValue("offset")
	if err == nil && offset_v != "" {
		offset, _ = strconv.Atoi(offset_v)
	}

	total, err = tr.File_ChildPagesMax(cui, path)
	if err != nil {
		H_error(cui, 500)
		return
	}

	paths, err := tr.File_ChildPagesList(cui, path, offset, total)
	if err != nil {
		H_error(cui, 500)
		return
	}

	FileUIApplyModOptsM(cui, paths)

	/* Output the page */
	type Page struct {
		*PfPage
		PageSize    int
		LastPage    int
		PagerOffset int
		PagerTotal  int
		Search      string
		Paths       []tr.PfFile
	}

	p := Page{cui.Page_def(), pageSize, tr.Template_Pager_LastPage(total, pageSize), offset, total, "", paths}
	cui.Page_show("file/list.tmpl", p)
}

func H_file_list_file(cui PfUI) {
	var m tr.PfFile
	var err error

	path := cui.GetSubPath()

	err = m.Fetch(cui, path, "")
	if err != nil {
		H_errmsg(cui, err)
		return
	}

	/* None HTML files are served directly */
	if m.MimeType != "text/html" {
		/* Cache for 30 minutes */
		cui.SetExpires(1 * 30)

		/* The file to serve */
		cui.SetStaticFile(m.FullFileName)

		/* The mime type */
		cui.SetContentType(m.MimeType)

		/* Done */
		return
	}

	/* HTML files are included in the page */
	type Page struct {
		*PfPage
		FileName string
	}

	p := Page{cui.Page_def(), m.FullFileName}
	cui.Page_show("file/view.tmpl", p)
}

func h_file_list(cui PfUI) {
	path := cui.GetSubPath()

	if tr.File_path_is_dir(path) {
		h_file_list_dir(cui)
		return
	}

	H_file_list_file(cui)
}

func h_file_details(cui PfUI) {
	var f tr.PfFile
	var err error

	path := cui.GetSubPath()

	err = f.Fetch(cui, path, "")
	if err != nil {
		cui.Dbgf("NOPE: %s: %s", path, err.Error())
		H_errmsg(cui, err)
		return
	}

	type move struct {
		Path     string `label:"New path of the file" pfreq:"yes"`
		Children bool   `label:"Move all children of this directory?" hint:"Only applies when the directory has children"`
		Confirm  bool   `label:"Confirm Moving" pfreq:"yes"`
		Action   string `label:"Action" pftype:"hidden"`
		Button   string `label:"Move" pftype:"submit"`
		Message  string /* Used by pfform() */
		Error    string /* Used by pfform() */
	}

	type del struct {
		Children bool   `label:"Delete all children of this directory?" hint:"Only applies when the directory has children"`
		Confirm  bool   `label:"Confirm Deletion" pfreq:"yes"`
		Action   string `label:"Action" pftype:"hidden"`
		Button   string `label:"Delete" pftype:"submit" htmlclass:"deny"`
		Message  string /* Used by pfform() */
		Error    string /* Used by pfform() */
	}

	type cpy struct {
		Path     string `label:"Path of the file/directory" pfreq:"yes"`
		Children bool   `label:"Copy all children of this directory?" hint:"Only applies when the directory has children"`
		Confirm  bool   `label:"Confirm copying" pfreq:"yes"`
		Action   string `label:"Action" pftype:"hidden"`
		Button   string `label:"Copy" pftype:"submit"`
		Message  string /* Used by pfform() */
		Error    string /* Used by pfform() */
	}

	/* TODO: Implement moving/copying files between groups */

	m := move{path, true, false, "move", "", "", ""}
	d := del{true, false, "delete", "", "", ""}
	c := cpy{path, true, false, "copy", "", "", ""}

	if cui.IsPOST() {
		action, err1 := cui.FormValue("action")
		confirmed, err2 := cui.FormValue("confirm")
		children, err3 := cui.FormValue("children")
		newpath, err4 := cui.FormValue("path")

		if err1 != nil || err2 != nil || err3 != nil {
			m.Error = "Invalid input"
			action = "Invalid"
		}

		if children == "on" {
			children = "yes"
		} else {
			children = "no"
		}

		switch action {
		case "move":
			if confirmed != "on" {
				m.Error = "Did not confirm"
			} else if err4 != nil {
				m.Error = "No Newpath"
			} else {
				mopts := tr.File_GetModOpts(cui)
				cmd := mopts.Cmdpfx + " move"
				arg := []string{path, newpath, children}

				_, err = cui.HandleCmd(cmd, arg)

				if err != nil {
					m.Error = err.Error()
				} else {
					opts := tr.File_GetModOpts(cui)
					url := tr.URL_Append(opts.URLpfx, opts.URLroot)
					url = tr.URL_Append(url, newpath)
					url += "?s=details"
					cui.SetRedirect(url, StatusSeeOther)
					return
				}
			}
			break

		case "delete":
			if confirmed != "on" {
				d.Error = "Did not confirm"
			} else {
				mopts := tr.File_GetModOpts(cui)
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

		case "copy":
			if confirmed != "on" {
				c.Error = "Did not confirm"
			} else if err4 != nil {
				m.Error = "No Newpath"
			} else {
				mopts := tr.File_GetModOpts(cui)
				cmd := mopts.Cmdpfx + " copy"
				arg := []string{path, newpath, children}

				_, err = cui.HandleCmd(cmd, arg)

				if err != nil {
					c.Error = err.Error()
				} else {
					url := "../"
					cui.SetRedirect(url, StatusSeeOther)
					return
				}
			}
			break

		case "Invalid":
			break

		default:
			m.Error = "Invalid action"
			break
		}
	}

	type Page struct {
		*PfPage
		File   tr.PfFile
		Move   move
		Delete del
		Copy   cpy
	}

	FileUIApplyModOpts(cui, &f)

	p := Page{cui.Page_def(), f, m, d, c}
	cui.Page_show("file/details.tmpl", p)
}

func h_file_add_dir(cui PfUI) {
	path := cui.GetSubPath()

	l := len(path)
	if l > 0 && path[l-1] != '/' {
		path += "/"
	}

	if cui.IsPOST() {
		path, err := cui.FormValue("name")
		desc, err2 := cui.FormValue("description")

		if err == nil && err2 == nil {
			/* Add a trailing / if the user didn't */
			pl := len(path)
			if pl == 0 || path[pl-1] != '/' {
				path += "/"
			}

			mopts := tr.File_GetModOpts(cui)
			cmd := mopts.Cmdpfx + " add_dir"
			arg := []string{path, desc}
			_, err = cui.HandleCmd(cmd, arg)

			if err == nil {
				opts := tr.File_GetModOpts(cui)
				url := tr.URL_Append(opts.URLpfx, opts.URLroot)
				url = tr.URL_Append(url, path)
				cui.SetRedirect(url, StatusSeeOther)
				return
			}
		} else {
			err = errors.New("Missing parameters")
		}

		H_errmsg(cui, err)
		return
	}

	type np struct {
		CurPath     string `label:"Current path:" pfset:"nobody" pfget:"user"`
		Name        string `label:"Filepath of new directory" pfreq:"yes" hint:"Can include '/' to create multiple sub-levels in one go"`
		Description string `label:"Description of new directory" pfreq:"yes" hint:"Short description"`
		Button      string `label:"Create new directory" pftype:"submit"`
	}

	type Page struct {
		*PfPage
		Opt np
	}

	p := Page{cui.Page_def(), np{path, path, "", ""}}
	cui.Page_show("file/add_dir.tmpl", p)
}

func h_file_add_file(cui PfUI) {
	path := cui.GetSubPath()

	if cui.IsPOST() {
		path, err1 := cui.FormValue("name")
		desc, err2 := cui.FormValue("description")
		file, fname, err := cui.GetFormFileReader("file")

		if err == nil && (err1 != nil || err2 != nil) {
			err = errors.New("Missing parameters")
		}

		if err != nil {
			H_errmsg(cui, err)
			return
		}

		/* Do we append the file name? */
		l := len(path)
		if l > 0 && path[l-1] == '/' {
			path += fname
		}

		/* Note: This avoids the CLI checks */
		err = tr.File_add_file(cui, path, desc, file)

		/* Close it */
		file.Close()

		if err == nil {
			/* Use the crumbpath here as we want the 'current' directory */
			cui.DelCrumb()
			url := cui.GetCrumbPath()
			cui.SetRedirect(url, StatusSeeOther)
			return
		}

		cui.Dbgf("File: FAILED adding: %s", err.Error())
		H_errmsg(cui, err)
		return
	}

	type np struct {
		CurPath     string `label:"Current path:" pfset:"nobody" pfget:"user"`
		Name        string `label:"File name" pfreq:"yes" hint:"Can include '/' to create multiple sub-levels in one go"`
		Description string `label:"Description of new file" pfreq:"yes" hint:"Short description"`
		File        string `label:"File" pfreq:"yes" pftype:"file" hint:"The File to upload"`
		Button      string `label:"Create new file" pftype:"submit"`
	}

	type Page struct {
		*PfPage
		Opt np
	}

	p := Page{cui.Page_def(), np{path, path, "", "", ""}}
	cui.Page_show("file/add_file.tmpl", p)
}

func file_edit_form(cui PfUI, path string) (err error) {
	mopts := tr.File_GetModOpts(cui)
	cmd := mopts.Cmdpfx + " update"
	arg := []string{path, path}

	_, err = cui.HandleCmd(cmd, arg)
	return
}

func H_file(cui PfUI) {
	/* URL of the page */
	cui.SetSubPath("/" + cui.GetPathString())
	for _, p := range cui.GetPath() {
		cui.AddCrumb(p, p, "")
	}

	sub := cui.GetArg("s")

	if sub == "list" {
		sub = ""
	}

	if sub != "" {
		sub = "?s=" + sub
	}

	ok, _ := cui.CheckPerms("H_file", PERM_USER)
	if ok {
		switch sub {
		case "?s=add_file":
			h_file_add_file(cui)
		case "?s=add_dir":
			h_file_add_dir(cui)
		case "?s=details":
			h_file_details(cui)
		default:
			h_file_list(cui)
		}
	} else {
		H_NoAccess(cui)
	}
}
