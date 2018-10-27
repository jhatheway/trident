package trident

import (
	"crypto/tls"
	"errors"
	"net/smtp"
)

const CRLF = "\r\n"

func Mail_PassResetUser(ctx PfCtx, email PfUserEmail, is_reset bool, nom_email PfUserEmail, user_portion string) (err error) {
	sys := System_Get()
	subject := ""
	body := "Dear " + email.FullName + "," + CRLF +
		CRLF

	if is_reset {
		subject = "Password Reset (User portion)"
		body += "A password reset request was made." + CRLF
	} else {
		subject = "New Account (User portion)"
		body += "Your account has been created. A new password thus has to be set." + CRLF
	}

	body += CRLF +
		"We are therefor sending you two token portions." + CRLF +
		"The user portion is in this email, the other portion " + CRLF +
		"has been sent to your nominator who will forward it in " + CRLF +
		"a secure method towards you." + CRLF +
		CRLF +
		"Your nominator is:" + CRLF +
		" " + nom_email.FullName + " <" + nom_email.Email + ">" + CRLF +
		CRLF +
		"When both parts have been received by you, please proceed to:" + CRLF +
		"  " + sys.PublicURL + "/recover/" +
		CRLF +
		"and enter the following password in the User Portion:" + CRLF +
		"  " + user_portion + CRLF +
		CRLF +
		"If you do not perform this reset the request will be canceled." + CRLF

	err = Mail(ctx,
		"", "",
		email.FullName, email.Email,
		true,
		subject,
		body,
		true,
		"",
		true)

	return
}

func Mail_PassResetNominator(ctx PfCtx, email PfUserEmail, is_reset bool, user_email PfUserEmail, nom_portion string) (err error) {
	subject := ""
	body := "Dear " + email.FullName + "," + CRLF +
		CRLF

	if is_reset {
		subject = "Password Reset (Nominator portion)"
		body += "A password reset request was made for:" + CRLF
	} else {
		subject = "New account (Nominator portion)"
		body += "A new account has been created for: " + CRLF
	}

	body += " " + user_email.FullName + " <" + user_email.Email + ">" + CRLF +
		CRLF +
		"As you are a nominator of this person, you are receiving " + CRLF +
		"the second portion of this email. " + CRLF +
		CRLF +
		"Please securely inform " + user_email.FullName + CRLF +
		"of the following Nominator Portion of the password reset: " + CRLF +
		"  " + nom_portion + CRLF

	err = Mail(ctx,
		"", "",
		email.FullName, email.Email,
		true,
		subject,
		body,
		true,
		"",
		true)

	return
}

/* TODO: Simple version, replace with internally queued edition later */
func mailA(ctx PfCtx, src_name string, src string, dst_name []string, dst []string, prefix bool, subject string, body string, regards bool, footer string, sysfooter bool) (err error) {
	if len(dst) != len(dst_name) {
		err = errors.New("Mismatch length in dst_name and dst options")
		return
	}

	sys := System_Get()

	server_host := Config.SMTP_host
	server_port := Config.SMTP_port
	server_ssl := Config.SMTP_SSL

	/* Default source? */
	if src_name == "" {
		src_name = sys.Name
	}

	/* Default source? */
	if src == "" {
		/*
		 * TODO: Special bounce address (bounce+flat_email@<mx>)
		 * then when mail returns status can be shown in the system
		 */
		src = "bounce@" + sys.EmailDomain
	}

	/* Apply <> to src + dst */
	src = "<" + src + ">"

	for d := range dst {
		dst[d] = "<" + dst[d] + ">"
	}

	/* Prefix Subject with Name? */
	if prefix {
		subject = "[" + sys.Name + "] " + subject
	}

	/* Add a nice regards */
	body += CRLF

	if regards {
		body += "Regards," + CRLF +
			"  " + sys.AdminName + " for " + sys.Name + CRLF
	}

	/* Add a footer showing system details? */
	if sysfooter || footer != "" {
		body += CRLF +
			"--" + CRLF
	}

	if sysfooter {
		body += sys.Name + " -- " + sys.PublicURL + CRLF
	}

	if footer != "" {
		if sysfooter {
			body += CRLF
		}

		body += footer
	}

	/* Connect to the local SMTP server */
	c, err := smtp.Dial(server_host + ":" + server_port)
	if err != nil {
		return
	}
	defer c.Close()

	/* Identify ourselves */
	err = c.Hello(Config.Nodename)
	if err != nil {
		return
	}

	/* Is there STARTTLS support? */
	starttls, _ := c.Extension("STARTTLS")
	if starttls {
		var tlsconfig *tls.Config

		/* Do require trust or ignore the certificate presented? */
		if server_ssl == "ignore" {
			tlsconfig = &tls.Config{InsecureSkipVerify: true}
		} else {
			tlsconfig = &tls.Config{ServerName: server_host}
		}

		/* Go for TLS */
		err = c.StartTLS(tlsconfig)
		if err != nil {
			return
		}
	}

	/* Set the sender and recipient */
	err = c.Mail(src)
	if err != nil {
		return
	}

	for d := range dst {
		err = c.Rcpt(dst[d])
		if err != nil {
			return
		}
	}

	/* Send the email body */
	w, err := c.Data()
	if err != nil {
		return
	}
	defer w.Close()

	headers := "From: " + "\"" + src_name + "\" " + src + CRLF

	for d := range dst {
		headers += "To: " + "\"" + dst_name[d] + "\" " + dst[d] + CRLF
	}

	headers +=
		"User-Agent: " + Config.UserAgent + CRLF +
			"Subject: " + subject + CRLF +
			CRLF

	w.Write([]byte(headers))
	w.Write([]byte(body))

	err = w.Close()
	if err != nil {
		return
	}

	/* Send the QUIT command and close the connection */
	err = c.Quit()
	if err != nil {
		return
	}

	return
}

/* Wrapper around the real mailA() function so we can handle errors in a single place */
func Mail(ctx PfCtx, src_name string, src string, dst_name string, dst string, prefix bool, subject string, body string, regards bool, footer string, sysfooter bool) (err error) {
	err = mailA(ctx, src_name, src, []string{dst_name}, []string{dst}, prefix, subject, body, regards, footer, sysfooter)
	if err != nil {
		ctx.Err("Sending email to " + dst + " failed: " + err.Error())
		err = errors.New("Sending email failed")
	}

	return
}

func MailM(ctx PfCtx, src_name string, src string, dst_name []string, dst []string, prefix bool, subject string, body string, regards bool, footer string, sysfooter bool) (err error) {
	err = mailA(ctx, src_name, src, dst_name, dst, prefix, subject, body, regards, footer, sysfooter)
	if err != nil {
		ctx.Err("Sending email failed: " + err.Error())
		err = errors.New("Sending email failed")
	}

	return
}

func Mail_VerifyEmail(ctx PfCtx, email PfUserEmail, verifycode string) (err error) {
	sys := System_Get()
	subject := "Email Verification Request"

	body := "Dear " + email.FullName + "," + CRLF +
		CRLF +
		"Somebody (probably you) has requested the email address:" + CRLF +
		"  " + email.Email + CRLF +
		"to be verified for " + sys.Name + " at " + sys.PublicURL + "." + CRLF +
		CRLF +
		"If you feel that this mail was sent to you without your consent, please" + CRLF +
		"reply to the administrator at:" + CRLF +
		"   " + sys.AdminName + " <" + sys.AdminEmail + ">" + CRLF +
		"and we will try to figure out what went wrong." + CRLF +
		CRLF +
		"To verify that this address is really yours, please visit the URL below" + CRLF +
		"and enter the token. This will ensure that you have read this mail and" + CRLF +
		"that your email address is valid." + CRLF +
		CRLF +
		"  " + sys.PublicURL +
		"/user/" + email.Member +
		"/email/" + email.Email +
		"/confirm/?verifycode=" + verifycode + CRLF +
		CRLF +
		"Or enter the verification code:" + CRLF +
		"  " + verifycode + CRLF +
		"in the interface for the email address " + email.Email + CRLF +
		CRLF +
		"If you do not verify this email address the request will be canceled." + CRLF

	err = Mail(ctx,
		"", "",
		email.FullName, email.Email,
		true,
		subject,
		body,
		true,
		"",
		true)

	return
}

func Mail_PasswordChanged(ctx PfCtx, email PfUserEmail) (err error) {
	sys := System_Get()
	subject := "Password changed"

	body := "Dear " + email.FullName + "," + CRLF +
		CRLF +
		"Somebody (probably you) has changed the password associated to your account:" + CRLF +
		"  " + email.Email + CRLF +
		CRLF +
		"If you did not change your password, please reply to the administrator at:" + CRLF +
		"   " + sys.AdminName + " <" + sys.AdminEmail + ">" + CRLF +
		"and we will try to figure out what went wrong." + CRLF

	err = Mail(ctx,
		"", "",
		email.FullName, email.Email,
		true,
		subject,
		body,
		true,
		"",
		true)

	return
}
