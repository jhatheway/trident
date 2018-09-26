#!/bin/bash

TGS="tg1 tg2 tg3 tg4"
USERS="user1 user2 user3 user4"
LISTS="warroom"

CLI=./src/cmd/tcli/tcli.go

echo "* Creating users..."
for US in ${USERS}
do
	echo "- ${US}"
	$CLI user new ${US} ${US}@example.org
	$CLI user set name_first ${US} ${US}
	$CLI user set name_last ${US} ${US}
	$CLI user set tel_info ${US} "+15552352352"
	$CLI user set sms_info ${US} "+15552352352"
	$CLI user set post_info ${US} "Some Street, The City"
	$CLI user set bio_info ${US} "So much to tell"
done

echo "* Creating TrustGroups..."
for TG in ${TGS};
do
	echo "- ${TG}"
	$CLI group add ${TG}
	$CLI group set descr ${TG} "Generated: ${TG}"

	for US in ${USERS}
	do
		echo " - adding user ${US}"
		$CLI group member add ${TG} ${US}
	done

	for ML in ${LISTS}
	do
		echo " ~ adding mailinglist ${ML}"
		$CLI ml new ${TG} ${ML}

		for US in ${USERS}
		do
			echo " = adding list member ${US}"
			$CLI ml member add ${TG} ${ML} ${US}
		done
	done
done

