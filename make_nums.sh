#!/bin/bash

for num in {1..4000}
do
    roman -a=true $num >> all_nums_add.txt
done