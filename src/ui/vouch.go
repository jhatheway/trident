package TriUI

import (
	"errors"
	"net/http"
	"strings"
	tr "lib"
)

func vouch_attestations_get(cui PfUI, grp tr.TriGroup) (attestations string, err error) {
	/* Get the group's attestations */
	required_attestations, err := grp.GetAttestations()
	if err != nil {
		return
	}

	/* The attentations in the form */
	fatts, err := cui.FormValueM("attestations")
	if err != nil {
		fatts, err = cui.FormValueM("attestations[]")
		if err != nil {
			err = errors.New("Missing attestations")
			return
		}
	}

	/* Walk required_attestations verify all are present */
	var atts []string
	for _, att := range required_attestations {

		/* Find it in the Form's values */
		found := false
		for _, fa := range fatts {
			if fa == att.Ident {
				found = true
				break
			}
		}

		if !found {
			err = errors.New("Incomplete attestations: " + att.Ident)
			return
		}

		atts = append(atts, att.Ident)
	}

	attestations = strings.Join(atts, ",")
	return
}

func vouch_args(cui PfUI) (err error) {
	vouchee, err := cui.FormValue("vouchee")
	if err != nil {
		return
	}

	groupname, err := cui.FormValue("group")
	if err != nil {
		return
	}

	err = cui.SelectGroup(groupname, tr.PERM_GROUP_MEMBER)
	if err != nil {
		return
	}

	tctx := tr.TriGetCtx(cui)

	/* Check member access to group */
	err = tctx.SelectVouchee(vouchee, tr.PERM_GROUP_MEMBER|tr.PERM_USER_VIEW)
	if err != nil {
		cui.Errf("Selecting Vouchee: %s", err.Error())
		H_error(cui, http.StatusNotFound)
		return
	}

	return
}

func vouch_edit(cui PfUI) (err error) {
	err = vouch_args(cui)
	if err != nil {
		return
	}

	tctx := tr.TriGetCtx(cui)
	grp := tctx.TriSelectedGroup()
	vouchee := tctx.SelectedVouchee()
	comment, err := cui.FormValue("comment")

	if err != nil {
		return
	}

	cmd := "group vouch update"
	arg := []string{grp.GetGroupName(), cui.TheUser().GetUserName(), vouchee.GetUserName(), comment}
	_, err = cui.HandleCmd(cmd, arg)

	return
}

func vouch_add(cui PfUI) (err error) {
	err = vouch_args(cui)
	if err != nil {
		return
	}

	tctx := tr.TriGetCtx(cui)

	grp := tctx.TriSelectedGroup()
	vouchee := tctx.SelectedVouchee()

	comment, err := cui.FormValue("comment")
	if err != nil {
		return
	}

	attestations, err := vouch_attestations_get(cui, grp)
	if err != nil {
		return
	}

	cmd := "group vouch add"
	arg := []string{grp.GetGroupName(), cui.TheUser().GetUserName(), vouchee.GetUserName(), comment, attestations}

	_, err = cui.HandleCmd(cmd, arg)

	return
}

func vouch_remove(cui PfUI) (err error) {
	err = vouch_args(cui)
	if err != nil {
		return
	}

	tctx := tr.TriGetCtx(cui)
	grp := tctx.TriSelectedGroup()
	vouchee := tctx.SelectedVouchee()

	cmd := "group vouch remove"
	arg := []string{grp.GetGroupName(), cui.TheUser().GetUserName(), vouchee.GetUserName()}

	_, err = cui.HandleCmd(cmd, arg)
	return
}

func vouch_nominate_new(cui PfUI) (msg string, err error) {
	tctx := tr.TriGetCtx(cui)

	var cmd string
	var args []string

	grp := tctx.TriSelectedGroup()

	email, err := cui.FormValue("email")
	descr, err2 := cui.FormValue("fullname")
	affil, err3 := cui.FormValue("affiliation")
	bio, err4 := cui.FormValue("biography")
	comment, err5 := cui.FormValue("comment")

	if err != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		err = errors.New("Invalid parameters provided")
		return
	}

	attestations, err6 := vouch_attestations_get(cui, grp)
	if err6 != nil {
		err = err6
		return
	}

	/* Generate a username */
	vouchee_ident, err := tr.Fullname_to_ident(descr)
	if err != nil {
		cui.Errf("Fullname to ident for %q failed: %s", err.Error())
		err = errors.New("Could not convert full name to ident")
		return
	}

	cmd = "user nominate"
	args = []string{vouchee_ident, email, bio, affil, descr}

	msg, err = cui.HandleCmd(cmd, args)
	if err != nil {
		return
	}

	cmd = "group nominate"
	args = []string{grp.GetGroupName(), vouchee_ident}

	msg, err = cui.HandleCmd(cmd, args)
	if err != nil {
		return
	}

	cmd = "group vouch add"
	args = []string{grp.GetGroupName(), cui.TheUser().GetUserName(), vouchee_ident, comment, attestations}

	msg, err = cui.HandleCmd(cmd, args)
	if err != nil {
		return
	}
	msg = "Member successfully nominated."
	return
}

func vouch_nominate(cui PfUI) (msg string, err error) {
	tctx := tr.TriGetCtx(cui)

	var cmd string
	var arg []string

	grp := tctx.TriSelectedGroup()
	vouchee := tctx.SelectedVouchee()

	comment, err := cui.FormValue("comment")
	if err != nil {
		err = errors.New("Invalid parameters")
		return
	}

	attestations, err := vouch_attestations_get(cui, grp)
	if err != nil {
		err = errors.New("Invalid attestation parameters")
		return
	}

	cmd = "group member add"
	arg = []string{grp.GetGroupName(), vouchee.GetUserName()}

	msg, err = cui.HandleCmd(cmd, arg)
	if err != nil {
		return
	}

	cmd = "group vouch add"
	arg = []string{grp.GetGroupName(), cui.TheUser().GetUserName(), vouchee.GetUserName(), comment, attestations}

	msg, err = cui.HandleCmd(cmd, arg)
	if err != nil {
		return
	}

	msg = "Nomination added"

	return
}
