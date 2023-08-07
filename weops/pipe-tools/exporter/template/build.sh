#!/bin/bash

for version in v1-16 v1-17 v1-18 v1-19 v1-20 v1-21 v1-22 v1-23 v1-24 v1-25; do
  standalone_output_file="standalone_${version}.yaml"
  sed "s/{{VERSION}}/${version}/g;" standalone.tpl > ../standalone/${standalone_output_file}
done
