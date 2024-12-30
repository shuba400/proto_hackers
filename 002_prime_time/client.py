import socket
import json

def test_server(host, port):
    def send_request(data):
        """Send a request to the server and get the response."""
        try:
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                print(f"Request: {data}")
                s.connect((host, port))
                s.sendall((json.dumps(data) + "\n").encode('utf-8'))
                response = s.recv(1024).decode('utf-8')
                if(not response.endswith('\n')):
                    print("No newline after response")
                print(f"Response: {response}\n")
                return response
        except Exception as e:
            print(f"Error: {e}\n")

    # Conforming request - prime number
    send_request({"method": "isPrime", "number": 972571858395545116790686578015198574111079184006423515688})

    # Conforming request - non-prime number
    send_request({"method": "isPrime", "number": -8})

    # Conforming request - floating-point number (not prime)
    send_request({"method": "isPrime", "number": 5.5})

    # Malformed request - missing "method" field
    send_request({"number": 7})

    # Malformed request - missing "number" field
    send_request({"method": "isPrime"})

    # Malformed request - incorrect "method" value
    send_request({"method": "checkPrime", "number": 7})

    # Malformed request - "number" is not a number
    send_request({"method": "isPrime", "number": "seven"})

    # Malformed request - not a JSON object
    send_request("Not a JSON object")

    # Conforming request with extraneous fields
    send_request({"method": "isPrime", "number": 7, "extra": "field"})

if __name__ == "__main__":
    HOST = "127.0.0.1"  # Replace with the server's IP
    PORT = 8080         # Replace with the server's port
    test_server(HOST, PORT)