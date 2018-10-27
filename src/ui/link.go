package TriUI

import (
	"html/template"
	tr "lib"
)

type PfLink struct {
	Link string
	Desc string
	Long string
	Subs []PfLink
}

func (l PfLink) HTML() (s template.HTML) {
	t := "<li>"

	t += "<a href=\"" + tr.HE(l.Link) + "\""

	if l.Long != "" {
		t += " title=\"" + tr.HE(l.Long) + "\""
	}

	t += ">" + tr.HE(l.Desc) + "</a>"

	if len(l.Subs) > 0 {
		t += "<ul>\n"
		for _, ll := range l.Subs {
			t += string(ll.HTML())
		}
		t += "</ul>\n"
	}

	t += "</li>\n"

	s = tr.HEB(t)
	return
}

type PfLinkCol struct {
	M []PfLink
}

func (c *PfLinkCol) Add(l PfLink) {
	c.M = append(c.M, l)
}

func (c *PfLinkCol) Pop() (l *PfLink) {
	l = nil
	ln := len(c.M)
	if ln > 0 {
		/* The last item */
		l = &c.M[ln-1]

		/* Remove that item from the list */
		c.M = c.M[:ln-1]
	}
	return
}

func (c *PfLinkCol) Len() int {
	return len(c.M)
}

func (c *PfLinkCol) Last() (l *PfLink) {
	l = nil
	ln := len(c.M)
	if ln > 0 {
		l = &c.M[ln-1]
	}
	return
}

func (c PfLinkCol) HTML(ul bool, class string) (s template.HTML) {
	if len(c.M) == 0 {
		return
	}

	if ul {
		s += "<ul"
		if class != "" {
			s += tr.HEB(" class=\"" + tr.HE(class) + "\"")
		}
		s += ">\n"
	}

	for _, l := range c.M {
		s += l.HTML()
	}

	if ul {
		s += "</ul>"
	}
	return
}
