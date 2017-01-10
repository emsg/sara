#/usr/bin/env python
#coding=utf8

import socket
import time
import uuid

self_jid = 'test2@a.a'

p = '''{"envelope":{"id":"1234567890","type":0,"jid":"%s","pwd":"abc123"},"vsn":"0.0.1"}\01'''
print p % self_jid

client = socket.socket(socket.AF_INET,socket.SOCK_STREAM)
client.connect(("127.0.0.1",4222))
client.send(p % self_jid)

try:
    print "recv:",client.recv(1024)
except :
    print "close"
print "bye bye"
client.close()
