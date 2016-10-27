#!/bin/bash

file_type=$1
file_name=$2


if [ $file_type = 'dir' ]; then 
	if [ -d "$file_name" ]; then 
	 echo 'true';
	else
	 echo 'false';
	fi
else
	if [ -f "$file_name" ]; then 
	 echo 'true';
	else
	 echo 'false';
	fi
fi

