package TriUI

import (
	"strconv"
	"strings"

	"trident.li/keyval"
	tr "lib"
)

func h_group_add(cui PfUI) {
	cmd := "group add"
	arg := []string{""}

	msg, err := cui.HandleCmd(cmd, arg)

	var errmsg = ""

	if err != nil {
		/* Failed */
		errmsg = err.Error()
	} else {
		group_name, _ := cui.FormValue("group")
		if group_name != "" {
			/* Success */
			cui.SetRedirect("/group/"+group_name+"/settings/", StatusSeeOther)
			return
		}
	}

	/* Output the page */
	type grpnew struct {
		Group  string `label:"Group Name" pfreq:"yes" hint:"The name of the group"`
		Button string `label:"Create" pftype:"submit"`
	}

	type Page struct {
		*PfPage
		Group   grpnew
		Message string
		Error   string
	}

	var grp grpnew
	p := Page{cui.Page_def(), grp, msg, errmsg}
	cui.Page_show("group/new.tmpl", p)
}

func h_group_settings(cui PfUI) {
	grp := cui.SelectedGroup()

	cmd := "group set"
	arg := []string{grp.GetGroupName()}

	msg, err := cui.HandleForm(cmd, arg, grp)

	var errmsg = ""

	if err != nil {
		/* Failed */
		errmsg = err.Error()
	} else {
		/* Success */
	}

	err = grp.Refresh()
	if err != nil {
		errmsg += err.Error()
	}

	/* Output the page */
	type Page struct {
		*PfPage
		Tg      tr.PfGroup
		Message string
		Error   string
	}

	p := Page{cui.Page_def(), grp, msg, errmsg}
	cui.Page_show("group/settings.tmpl", p)
}

func h_group_log(cui PfUI) {
	grp := cui.SelectedGroup()
	h_system_logA(cui, "", grp.GetGroupName())
}



func h_group_cmd(cui PfUI) {
	grp := cui.SelectedGroup()

	username, err := cui.FormValue("user")
	if err != nil {
		cui.Errf("Missing parameter user: %s", err.Error())
		return
	}

	groupname, err := cui.FormValue("group")
	if err != nil {
		cui.Errf("Missing parameter group: %s", err.Error())
		return
	}

	if grp.GetGroupName() != groupname {
		cui.Errf("Mismatching group %q vs %q", grp.GetGroupName(), groupname)
		return
	}

	cmd, err := cui.FormValue("cmd")
	if err != nil {
		cui.Errf("Missing parameter cmd: %s", err.Error())
		return
	}

	err = cui.SelectUser(username, PERM_GROUP_ADMIN)
	if err != nil {
		cui.Errf("Could not select user %q: %s", username, err.Error())
		return
	}

	user := cui.SelectedUser()

	switch cmd {
	case "block":
	case "unblock":
	case "promote":
	case "demote":
	case "approve":
	default:
		cui.Errf("Unknown Group command: %q", cmd)
		return
	}

	cmd = "group member " + cmd

	/* The arguments */
	arg := []string{grp.GetGroupName(), user.GetUserName()}

	_, err = cui.HandleCmd(cmd, arg)
	if err != nil {
		cui.Err(err.Error())
		return
	}

	cui.SetRedirect("/group/"+grp.GetGroupName()+"/member/", StatusSeeOther)
	return
}

func h_group_index(cui PfUI) {

	/* Output the page */
	type Page struct {
		*PfPage
		GroupName string
		GroupDesc string
	}

	grp := cui.SelectedGroup()

	p := Page{cui.Page_def(), grp.GetGroupName(), grp.GetGroupDesc()}
	cui.Page_show("group/index.tmpl", p)
}

func h_group_list(cui PfUI) {
	grp := cui.NewGroup()
	var grusers []tr.PfGroupMember
	var err error

	if !cui.IsSysAdmin() {
		grusers, err = grp.GetGroups(cui, cui.TheUser().GetUserName())
	} else {
		grusers, err = grp.GetGroupsAll()
	}

	if err != nil {
		return
	}

	grps := make(map[string]string)
	for _, gru := range grusers {
		if cui.IsSysAdmin() || gru.GetGroupCanSee() {
			grps[gru.GetGroupName()] = gru.GetGroupDesc()
		}
	}

	/* Output the page */
	type Page struct {
		*PfPage
		Groups map[string]string
	}

	menu := NewPfUIMenu([]PfUIMentry{
		{"add", "Add Group", PERM_GROUP_ADMIN, h_group_add, nil},
	})

	cui.SetPageMenu(&menu)

	p := Page{cui.Page_def(), grps}
	cui.Page_show("group/list.tmpl", p)
}

func H_group_member_profile(cui PfUI) {
	path := cui.GetPath()

	/* Select the user */
	err := cui.SelectUser(path[0], PERM_USER_VIEW)
	if err != nil {
		cui.Err("User: " + err.Error())
		H_NoAccess(cui)
		return
	}

	h_user(cui)
	return
}

func h_group_pgp_keys(cui PfUI) {
	var output []byte
	keyset := make(map[[16]byte][]byte)
	grp := cui.SelectedGroup()

	err := grp.GetKeys(cui, keyset)
	if err != nil {
		/* Temp redirect to unknown */
		H_error(cui, StatusNotFound)
		return
	}

	fname := grp.GetGroupName() + ".asc"

	for k := range keyset {
		output = append(output, keyset[k][:]...)
		output = append(output, byte(0x0a))
	}

	cui.SetContentType("application/pgp-keys")
	cui.SetFileName(fname)
	cui.SetExpires(60)
	cui.SetRaw(output)
}

func h_group_airports(cui PfUI) {
	iata := cui.GetArg("iata")

	grp := cui.SelectedGroup()

	members, err := grp.ListGroupMembers("", "", 0, 0, false, false, false)
	if err != nil {
		H_errmsg(cui, err)
		return
	}

	airports := make(map[string]int)
	for _, m := range members {
		airport := m.GetAirport()
		_, ok := airports[airport]
		if !ok {
			airports[airport] = 0
		}

		airports[airport]++
	}

	type Airport struct {
		IATA  string
		Count int
	}

	var as []Airport

	found := true

	for k, n := range airports {
		airport := Airport{k, n}
		as = append(as, airport)

		/* Found the airport we are looking for? */
		if k == iata {
			found = true
		}
	}

	/* Search is invalid */
	if !found {
		iata = ""
	}

	/* Remove members that are not at the selected airport */
	removed := 0
	for a := range members {
		j := a - removed
		if members[j].GetAirport() != iata {
			members = members[:j+copy(members[j:], members[j+1:])]
			removed++
		}
	}

	/* Output the page */
	type Page struct {
		*PfPage
		Airport  string
		Airports []Airport
		Members  []tr.PfGroupMember
	}

	p := Page{cui.Page_def(), iata, as, members}
	cui.Page_show("group/airports.tmpl", p)
}

func h_group_languages(cui PfUI) {
	langSearchCode := cui.GetArg("language")
	langSearchName := ""
	grp := cui.SelectedGroup()

	members, err := grp.ListGroupMembers("", "", 0, 0, false, false, false)
	if err != nil {
		H_errmsg(cui, err)
		return
	}

	countedLanguages := make(map[tr.PfLanguage]int)
	membersWhoUnderstand := make(map[tr.PfGroupMember]string)
	/* First we need to go over all the group members, getting all the languages they know */
	for _, m := range members {
		languages, err := tr.GetUserLanguages(m.GetUserName())
		if err != nil {
			err.Error()
		} else {
			/* Then for each group member, count all the languages */
			for _, l := range languages {
				_, ok := countedLanguages[l.Language]
				if !ok {
					countedLanguages[l.Language] = 0
				}
				countedLanguages[l.Language]++

				/* For the given language, if it matches what the user was filtering on (if
				   a filter was given), then add that member to the list that we pass back */
				if l.Language.Code == langSearchCode {
					membersWhoUnderstand[m] = l.Skill
					langSearchName = l.Language.Name
				}
			}
		}
	}

	/* Output the page */
	type Page struct {
		*PfPage
		Language  string
		Languages map[tr.PfLanguage]int
		Members   map[tr.PfGroupMember]string
	}

	p := Page{cui.Page_def(), langSearchName, countedLanguages, membersWhoUnderstand}
	cui.Page_show("group/languages.tmpl", p)
}

func h_group_contacts_vcard(cui PfUI) {
	grp := cui.SelectedGroup()

	vcard, err := grp.GetVcards()
	if err != nil {
		H_errmsg(cui, err)
		return
	}

	fname := grp.GetGroupName() + ".vcf"

	cui.SetContentType("text/vcard")
	cui.SetFileName(fname)
	cui.SetExpires(60)
	cui.SetRaw([]byte(vcard))
	return
}

func h_group_contacts(cui PfUI) {
	fmt := cui.GetArg("format")

	if fmt == "vcard" {
		h_group_contacts_vcard(cui)
		return
	}

	grp := cui.SelectedGroup()

	members, err := grp.ListGroupMembers("", "", 0, 0, false, false, false)
	if err != nil {
		H_errmsg(cui, err)
		return
	}

	/* Output the page */
	type Page struct {
		*PfPage
		Members []tr.PfGroupMember
	}

	p := Page{cui.Page_def(), members}
	cui.Page_show("group/contacts.tmpl", p)
}

func h_group_file(cui PfUI) {
	/* Module options */
	tr.Group_FileMod(cui)

	/* Call the module */
	H_file(cui)
}

func h_group_wiki(cui PfUI) {
	/* Module options */
	tr.Group_WikiMod(cui)

	/* Call the module */
	H_wiki(cui)
}
func h_group_member(cui PfUI) {
	path := cui.GetPath()
	pageSize := tr.PAGER_PERPAGE /* TODO: Eventually I'd like this to come in from a parameter */

	if len(path) != 0 && path[0] != "" {
		H_group_member_profile(cui)
		return
	}

	var err error

	tctx := tr.TriGetCtx(cui)
	total := 0
	offset := 0

	offset_v, err := cui.FormValue("offset")
	if err == nil && offset_v != "" {
		offset, _ = strconv.Atoi(offset_v)
	}

	search, err := cui.FormValue("search")
	if err != nil {
		search = ""
	}

	grp := tctx.TriSelectedGroup()

	total, err = grp.ListGroupMembersTot(search)
	if err != nil {
		cui.Err("error: " + err.Error())
		return
	}

	members, err := grp.ListGroupMembers(search, cui.TheUser().GetUserName(), offset, 10, false, cui.IAmGroupAdmin(), false)
	if err != nil {
		cui.Err(err.Error())
		return
	}

	/* Output the page */
	type Page struct {
		*PfPage
		Group        tr.PfGroup
		GroupMembers []tr.PfGroupMember
		PageSize     int
		LastPage     int
		PagerOffset  int
		PagerTotal   int
		Search       string
		IsAdmin      bool
	}
	isadmin := cui.IAmGroupAdmin()
	p := Page{cui.Page_def(), grp, members, pageSize, tr.Template_Pager_LastPage(total, pageSize), offset, total, search, isadmin}
	cui.Page_show("group/members.tmpl", p)
}

type NominateAdd struct {
	group        tr.TriGroup
	Action       string          `label:"Action" pftype:"hidden"`
	Vouchee      string          `label:"Username" pfset:"nobody" pfget:"none"`
	Comment      string          `label:"Vouch comment" pftype:"text" hint:"Vouch description for this user" pfreq:"yes"`
	Attestations map[string]bool `label:"Attestations (all required)" hint:"Attestations for this user" options:"GetAttestationOpts" pfcheckboxmode:"yes"`
	Button       string          `label:"Nominate" pftype:"submit"`
	Message      string          /* Used by pfform() */
	Error        string          /* Used by pfform() */
}

func (na *NominateAdd) GetAttestationOpts(obj interface{}) (kvs keyval.KeyVals, err error) {
	return na.group.GetAttestationsKVS()
}

func h_group_nominate_existing(cui PfUI) {
	msg := ""
	errmsg := ""
	tctx := tr.TriGetCtx(cui)
	grp := tctx.TriSelectedGroup()

	vouchee_name, err := cui.FormValue("vouchee")
	if err != nil {
		H_errtxt(cui, "No valid vouchee")
		return
	}

	err = tctx.SelectVouchee(vouchee_name, PERM_USER_NOMINATE)
	if err != nil {
		H_errtxt(cui, "Vouchee unselectable")
		return
	}

	if cui.IsPOST() {
		action, err := cui.FormValue("action")
		if err == nil && action == "nominate" {
			msg, err = vouch_nominate(cui)
			if err != nil {
				errmsg = err.Error()
			}
		}
	}

	vouchee := tctx.SelectedVouchee()

	type Page struct {
		*PfPage
		Vouchee     string
		GroupName   string
		NominateAdd *NominateAdd
	}

	na := &NominateAdd{group: grp, Vouchee: vouchee.GetUserName(), Action: "nominate", Message: msg, Error: errmsg}

	p := Page{cui.Page_def(), vouchee.GetUserName(), grp.GetGroupName(), na}
	cui.Page_show("group/nominate_existing.tmpl", p)
}

type NominateNew struct {
	group        tr.TriGroup
	Action       string          `label:"Action" pftype:"hidden"`
	Search       string          `label:"Search" pftype:"hidden"`
	Email        string          `label:"Email address of nominee" pfset:"none"`
	FullName     string          `label:"Full Name" hint:"Full Name of this user" pfreq:"yes"`
	Affiliation  string          `label:"Affiliation" hint:"Who the user is affiliated to" pfreq:"yes"`
	Biography    string          `label:"Biography" pftype:"text" hint:"Biography for this user" pfreq:"yes"`
	Comment      string          `label:"Vouch Comment" pftype:"text" hint:"Vouch for this user" pfreq:"yes"`
	Attestations map[string]bool `label:"Attestations (all required)" hint:"Attestations for this user" options:"GetAttestationOpts" pfcheckboxmode:"yes"`
	Button       string          `label:"Nominate" pftype:"submit"`
	Message      string          /* Used by pfform() */
	Error        string          /* Used by pfform() */
}

func (na *NominateNew) GetAttestationOpts(obj interface{}) (kvs keyval.KeyVals, err error) {
	return na.group.GetAttestationsKVS()
}

func h_group_nominate(cui PfUI) {
	var msg string
	var err error
	var errmsg string
	var list []tr.PfUser
	var search string

	tctx := tr.TriGetCtx(cui)
	user := tctx.TriSelectedUser()
	grp := tctx.TriSelectedGroup()
	added := false

	/* Something posted? */
	if cui.IsPOST() {
		/* An action to perform? */
		action, err := cui.FormValue("action")
		if err == nil && action == "nominate" {
			msg, err = vouch_nominate_new(cui)
			if err != nil {
				errmsg += err.Error()
			}
			added = true
		}

		/* Search field? */
		search, err = cui.FormValue("search")
		if err != nil {
			search = ""
		}

		/* case-fold to lowercase */
		search = strings.ToLower(search)

		/* Simple 'is there an @ sign, it must be an email address' check */
		if strings.Index(search, "@") == -1 {
			/* Not an email, do not allow searching */
			search = ""
		}
	}

	/* Need to search the list? */
	notfound := true
	if search != "" {
		/* Get list of users matching the given search query */
		list, err = user.GetList(cui, search, 0, 0, true)
		if err != nil {
			cui.Errf("Listing users failed: %s", err.Error())
			H_error(cui, StatusBadRequest)
			return
		}

		if len(list) != 0 {
			notfound = false
		}
	}

	type Page struct {
		*PfPage
		Search    string
		GroupName string
		Users     []tr.PfUser
		NotFound  bool
		NewForm   *NominateNew
	}

	if added {
		notfound = true
	}

	/* Re-fill in the form (for people who do not enable the attestations) */
	descr, _ := cui.FormValue("fullname")
	affil, _ := cui.FormValue("affiliation")
	bio, _ := cui.FormValue("biography")
	comment, _ := cui.FormValue("comment")

	newform := &NominateNew{group: grp, Action: "nominate", Email: search, Message: msg, Error: errmsg, Search: search, FullName: descr, Affiliation: affil, Biography: bio, Comment: comment}

	p := Page{cui.Page_def(), search, grp.GetGroupName(), list, notfound, newform}
	cui.Page_show("group/nominate.tmpl", p)
}

func h_vouches_csv(cui PfUI) {
	grp := cui.SelectedGroup()

	vouches, err := tr.Vouches_Get(cui, grp.GetGroupName())
	if err != nil {
		H_errmsg(cui, err)
		return
	}

	csv := ""

	for _, v := range vouches {
		csv += v.Vouchor + "," + v.Vouchee + "," + v.Entered.Format(tr.Config.DateFormat) + "\n"
	}

	fname := grp.GetGroupName() + ".csv"

	cui.SetContentType("text/vcard")
	cui.SetFileName(fname)
	cui.SetExpires(60)
	cui.SetRaw([]byte(csv))
	return
}

func h_vouches(cui PfUI) {
	fmt := cui.GetArg("format")

	if fmt == "csv" {
		h_vouches_csv(cui)
		return
	}

	grp := cui.SelectedGroup()
	vouches, err := tr.Vouches_Get(cui, grp.GetGroupName())
	if err != nil {
		H_errmsg(cui, err)
		return
	}

	/* Output the page */
	type Page struct {
		*PfPage
		Vouches []tr.Vouch
	}

	p := Page{cui.Page_def(), vouches}
	cui.Page_show("group/vouches.tmpl", p)
}

func h_group(cui PfUI) {
	path := cui.GetPath()

	if len(path) == 0 || path[0] == "" {
		cui.SetPageMenu(nil)
		h_group_list(cui)
		return
	}

	/* New group creation */
	if path[0] == "add" && cui.IsSysAdmin() {
		cui.AddCrumb(path[0], "Add Group", "Add Group")
		cui.SetPageMenu(nil)
		h_group_add(cui)
		return
	}

	/* Check member access to group */
	err := cui.SelectGroup(cui.GetPath()[0], PERM_GROUP_MEMBER)
	if err != nil {
		cui.Err("Group: " + err.Error())
		H_NoAccess(cui)
		return
	}

	grp := cui.SelectedGroup()

	cui.AddCrumb(path[0], grp.GetGroupName(), grp.GetGroupDesc())

	cui.SetPath(path[1:])

	/* /group/<grp>/{path} */
	menu := NewPfUIMenu([]PfUIMentry{
		{"", "", PERM_GROUP_MEMBER, h_group_index, nil},
		{"settings", "Settings", PERM_GROUP_ADMIN, h_group_settings, nil},
		{"member", "Members", PERM_GROUP_MEMBER, h_group_member, nil},
		{"pgp_keys", "PGP Keys", PERM_GROUP_MEMBER, h_group_pgp_keys, nil},
		{"airports", "Airports", PERM_GROUP_MEMBER, h_group_airports, nil},
		{"languages", "Languages", PERM_GROUP_MEMBER, h_group_languages, nil},
		{"ml", "Mailing List", PERM_GROUP_MEMBER, h_ml, nil},
		{"wiki", "Wiki", PERM_GROUP_WIKI, h_group_wiki, nil},
		{"log", "Audit Log", PERM_GROUP_ADMIN, h_group_log, nil},
		{"file", "Files", PERM_GROUP_FILE, h_group_file, nil},
		{"contacts", "Contacts", PERM_GROUP_MEMBER, h_group_contacts, nil},
		{"cmd", "Commands", PERM_GROUP_ADMIN | PERM_HIDDEN | PERM_NOCRUMB, h_group_cmd, nil},
		{"nominate", "Nominate", PERM_GROUP_MEMBER, h_group_nominate, nil},
		{"nominate_existing", "Nominate existing user", PERM_GROUP_MEMBER | PERM_HIDDEN, h_group_nominate_existing, nil},
		{"vouches", "Vouches", PERM_GROUP_MEMBER, h_vouches, nil},
		{"vcp", "Vouching Control Panel", PERM_GROUP_MEMBER, h_group_vcp, nil},
		// TODO: {"calendar", "Calendar", PERM_GROUP_CALENDAR, h_calendar},
	})

	cui.UIMenu(menu)
}
