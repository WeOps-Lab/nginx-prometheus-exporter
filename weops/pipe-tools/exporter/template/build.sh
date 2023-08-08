#!/bin/bash

for version in v1-16 v1-17 v1-18 v1-19 v1-20 v1-21 v1-22 v1-23 v1-24 v1-25; do
  for module in stub vts; do
    standalone_output_file="${module}_${version}.yaml"
    sed "s/{{VERSION}}/${version}/g;s/{{MODULE}}/${module}/g" standalone.tpl > ../standalone/${standalone_output_file}
  done
done
