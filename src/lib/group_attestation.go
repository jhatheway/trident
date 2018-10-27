package trident

import (
	"trident.li/keyval"
)

type TriGroupAttestation struct {
	Ident      string
	Descr      string
	TrustGroup string
}

func (grpa *TriGroupAttestation) toString() (out string) {
	out = "{" + grpa.Ident + "} " + grpa.Descr
	return
}

func (grp *TriGroupS) GetAttestations() (output []TriGroupAttestation, err error) {
	q := "SELECT at.ident, " +
		"at.descr, " +
		"at.trustgroup " +
		"FROM attestations at " +
		"WHERE at.trustgroup = $1"
	rows, err := DB.Query(q, grp.GetGroupName())
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var at TriGroupAttestation
		err = rows.Scan(&at.Ident, &at.Descr, &at.TrustGroup)
		if err != nil {
			return
		}

		output = append(output, at)
	}

	return
}

func (grp *TriGroupS) GetAttestationsKVS() (kvs keyval.KeyVals, err error) {
	attestations, err := grp.GetAttestations()
	if err != nil {
		return
	}

	for _, att := range attestations {
		kvs.Add(att.Ident, att.Descr)
	}

	return
}

func grp_attestation_list(ctx PfCtx, args []string) (err error) {
	groupname := args[0]

	tctx := TriGetCtx(ctx)

	err = ctx.SelectGroup(groupname, PERM_GROUP_MEMBER)
	if err != nil {
		return
	}

	grp := tctx.TriSelectedGroup()

	ats, err := grp.GetAttestations()
	var at TriGroupAttestation
	for _, at = range ats {
		ctx.OutLn(at.toString())
	}

	return
}

func grp_attestation_add(ctx PfCtx, args []string) (err error) {
	groupname := args[0]
	ident := args[1]
	descr := args[2]

	err = ctx.SelectGroup(groupname, PERM_GROUP_ADMIN)
	if err != nil {
		return
	}

	tctx := TriGetCtx(ctx)
	grp := tctx.TriSelectedGroup()

	q := "INSERT INTO attestations " +
		"(ident, descr, trustgroup) " +
		"VALUES($1, $2, $3)"
	err = DB.Exec(ctx,
		"Added new attestation ($1,$2,$3)",
		1, q,
		ident, descr, grp.GetGroupName())
	if err != nil {
		return
	}

	return
}

func grp_attestation_delete(ctx PfCtx, args []string) (err error) {
	groupname := args[0]
	ident := args[1]
	descr := args[2]

	err = ctx.SelectGroup(groupname, PERM_GROUP_ADMIN)
	if err != nil {
		return
	}

	tctx := TriGetCtx(ctx)
	grp := tctx.TriSelectedGroup()

	q := "DELETE FROM attestations " +
		"WHERE ident = $1 " +
		"AND descr = $2 " +
		"AND trustgroup = $3"
	err = DB.Exec(ctx,
		"Removed attestation ($1,$2,$3)",
		1, q,
		ident, descr, grp.GetGroupName())
	if err != nil {
		return
	}

	return
}

func attestation_menu(ctx PfCtx, args []string) (err error) {
	menu := NewPfMenu([]PfMEntry{
		{"list", grp_attestation_list, 1, 1, []string{"groupname"}, PERM_USER, "List attestations"},
		{"add", grp_attestation_add, 3, 3, []string{"groupname", "ident", "Description"}, PERM_GROUP_ADMIN, "Add an attestation"},
		{"delete", grp_attestation_delete, 2, 2, []string{"groupname", "ident"}, PERM_GROUP_ADMIN, "Delete an attestation"},
	})

	err = ctx.Menu(args, menu)
	return
}
