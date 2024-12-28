import socket
import threading
import random
import string

# Function to generate a random message
def generate_random_message(length=200000):
    return ''.join(random.choices(string.ascii_letters + string.digits, k=length))
 
# Function to handle a single TCP connection
def handle_connection(server_ip, server_port, message_id):
    try:
        client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        client_socket.connect((server_ip, server_port))

        message = generate_random_message()
        print(f"[Message {message_id}] Sending: {message}")
        client_socket.sendall(message.encode()) 
        client_socket.shutdown(socket.SHUT_WR) # This to make sure that we actually do sent an EOF signal, otherwise server will keep on waiting for data

        response = b""
        while True:
            data = client_socket.recv(1024)
            if not data:  # No more data, connection closed by server
                break
            response += data

        print(f"[Message {message_id}] Received: {response.decode()}")
        if(message != response.decode()):
            print("Not same :")
        else:
            print("Same")

        # Close the connection
        client_socket.close()
    except Exception as e:
        # Close the socket
        client_socket.close()
        print(f"[Message {message_id}] Error: {e}")

def fire_requests(server_ip, server_port, num_requests):
    threads = []

    for i in range(num_requests):
        thread = threading.Thread(target=handle_connection, args=(server_ip, server_port, i + 1))
        threads.append(thread)
        thread.start()

    # Wait for all threads to complete
    for thread in threads:
        thread.join()

if __name__ == "__main__":
    # Server details
    SERVER_IP = "127.0.0.1"  # Replace with the actual server IP
    SERVER_PORT = 8080       # Replace with the actual server port

    # Number of requests to send
    NUM_REQUESTS = 1

    fire_requests(SERVER_IP, SERVER_PORT, NUM_REQUESTS)