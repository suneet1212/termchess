import socket
# Source - https://stackoverflow.com/a/18746406
# Posted by Manoj Pandey
# Retrieved 2026-09-25, License - CC BY-SA 3.0

UDP_IP = "2a02:aa11:b180:2300:55ee:beb2:9e3e:aa6b"
UDP_PORT = 40900
MESSAGE =  "Hello from Suneet"
if __name__ == "__main__":
    print("UDP target IP:", UDP_IP)
    print("UDP target port:", UDP_PORT)
    print("message:",MESSAGE)

    sock = socket.socket(socket.AF_INET6, socket.SOCK_DGRAM) # UDP
    sock.bind(("", UDP_PORT))
    sock.sendto(bytes(MESSAGE, "utf-8"), (UDP_IP, UDP_PORT))
