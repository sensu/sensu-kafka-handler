#!/bin/bash
SRC="secrets/kafka.producer.keystore.jks"
DES="secrets/kafka.producer.keystore.pkcs12"
DESPASS=dummy1234
PEM="tls/producer.pem"
KEY="tls/producer-key.pem"
CRT="tls/producer.crt"
CLIENT=producer

keytool -alias $CLIENT -importkeystore -srckeystore $SRC -destkeystore $DES -srcstoretype jks -deststoretype pkcs12 -destkeypass $DESPASS 

openssl pkcs12 -in $DES -out $PEM

openssl x509 -outform der -in $PEM -out $CRT

openssl rsa -in $PEM -out $KEY

