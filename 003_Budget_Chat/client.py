import socket
import threading

def handle_server_messages(client_socket):
    while True:
        try:
            message = client_socket.recv(1024).decode('utf-8')
            if message:
                print(f"{message.strip()}")
            else:
                print("Server disconnected.")
                break
        except Exception as e:
            print(f"Error receiving message: {e}")
            break

def main():
    host = '127.0.0.1'  
    port = 8080        

    try:
        client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        client_socket.connect((host, port))
        print("Connected to the server.")

        threading.Thread(target=handle_server_messages, args=(client_socket,), daemon=True).start()

        # Main loop to send messages to the server
        while True:
            try:
                message = input()
                if message.lower() == 'exit':
                    print("Disconnecting...")
                    break
                client_socket.sendall((message + "\n").encode('utf-8'))
            except Exception as e:
                print(f"Error sending message: {e}")
                break
    except ConnectionRefusedError:
        print("Unable to connect to the server. Ensure the server is running and accessible.")
    except Exception as e:
        print(f"An error occurred: {e}")
    finally:
        client_socket.close()
        print("Connection closed.")

if __name__ == "__main__":
    main()
