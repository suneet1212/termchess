import socket
import time

UDP_IP = "10.0.1.2"
UDP_PORT = 40000
MESSAGE = "hello p1!"

print("UDP target IP:", UDP_IP)
print("UDP target port:", UDP_PORT)
print("message:", MESSAGE)

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM) # UDP
sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)

sock.bind(("0.0.0.0", UDP_PORT))
for _ in range(10):
    sock.sendto(bytes(MESSAGE, "utf-8"), (UDP_IP, UDP_PORT))
    time.sleep(0.5)
sock.close()

