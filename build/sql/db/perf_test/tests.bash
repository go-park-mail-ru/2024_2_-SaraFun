#!/bin/bash

echo "Запуск makeData.go..."
go run makeData.go

if [ $? -ne 0 ]; then
    echo "Ошибка при выполнении makeData.go"
    exit 1
fi

echo "Запуск wrk для тестирования API..."
wrk -t1 -c1 -d600s -s script.lua https://localhost:8080/api/auth/signup

if [ $? -ne 0 ]; then
    echo "Ошибка при выполнении wrk"
    exit 1
fi

echo "Все команды выполнены успешно."