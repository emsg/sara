#/usr/bin/env python
#coding=utf8

import socket
import time
import uuid

p = '''{"envelope":{"id":"%s","type":0,"jid":"test1@a.a","pwd":"abc123"},"vsn":"0.0.1"}\01'''

client = socket.socket(socket.AF_INET,socket.SOCK_STREAM)
client.connect(("127.0.0.1",4222))
p = p % uuid.uuid4().hex
client.sendall(p.encode('utf-8'))

try:
    print "recv:",client.recv(1024)
    while True:
        time.sleep(6)
        print "👮 send: ❤️"
        client.sendall("\02\01".encode('utf-8'))
        print "recv:",client.recv(1024)
        print "----------------------"
except :
    print "close"
print "bye bye"
client.close()
