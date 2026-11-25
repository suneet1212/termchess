import socket

UDP_IP = "127.0.0.1"
UDP_PORT = 18745

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
sock.bind((UDP_IP, UDP_PORT))

while True:
    data, addr = sock.recvfrom(1024)
    print(f"From {addr}: {data}")
    if len(data) > 0:
        sock.sendto(bytes("hello", "utf-8"), (addr[0], addr[1]))
