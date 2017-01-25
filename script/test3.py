#/usr/bin/env python
#coding=utf8

import socket
import time
import uuid

p = '''{"envelope":{"id":"%s","type":0,"jid":"1@a.a","pwd":"abc123"},"vsn":"0.0.1"}\01'''

m = '''{"envelope":{"id":"%s","type":2,"from":"1@a.a","gid":"4"},"payload":{"content":"hi all"},"vsn":"0.0.1"}\01'''

client = socket.socket(socket.AF_INET,socket.SOCK_STREAM)
client.connect(("127.0.0.1",4222))
p = p % uuid.uuid4().hex
client.sendall(p.encode('utf-8'))

try:
    print "recv:",client.recv(1024)
    for i in range(0,1) :
        time.sleep(3)
        print "👮 send: ❤️"
        client.sendall(m % uuid.uuid4().hex)
        print "recv:",client.recv(1024)
        print "----------------------",i
except :
    print "close"
print "bye bye"
client.close()
