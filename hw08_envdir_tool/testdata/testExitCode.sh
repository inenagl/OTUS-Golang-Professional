#!/usr/bin/env bash

# Берём код выхода либо из переменной окружения EXIT_CODE, либо из первого аргумента командной строки, либо 0
if [ -z "${EXIT_CODE}" ]; then
  if [ $# -gt 0 ]; then
    code=$1
  else
    code=0
  fi
else
  code=${EXIT_CODE}
fi

exit $code