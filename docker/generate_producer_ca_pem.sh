#!/bin/bash
SRC="secrets/kafka.producer.kestore.jks"
DES="secrets/kafka.producer.keystore.pkcs12"
PEM="tls/producer_ca.pem"

keytool -alias caroot -importkeystore -srckeystore $SRC -destkeystore $DES -srcstoretype jks -deststoretype pkcs12  

openssl pkcs12 -in $DES -out $PEM

#openssl x509 -outform der -in $PEM -out $CRT
#
#openssl rsa -in $PEM -out $KEY

