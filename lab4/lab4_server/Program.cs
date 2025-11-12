using System.Net;
using System.Net.Sockets;
using System.Text;

namespace lab4_server
{
    class Program
    {

        static async Task Main(string[] args)
        {

            var listener = new TcpListener(IPAddress.Loopback, 8080);
            listener.Start();
            Console.WriteLine("Waiting for connections...");
            while (true)
            {
                var client = listener.AcceptTcpClient();
                _ = Task.Run(() => HandleClient(client));
            }
        }


        static void HandleClient(TcpClient client)
        {
            try
            {
                using var stream = client.GetStream();
                var buffer = new byte[4096];
                int bytesRead = stream.Read(buffer, 0, buffer.Length);
                string request = Encoding.UTF8.GetString(buffer, 0, bytesRead);
                Console.WriteLine($"[Server] Received request:\n{request}");

                // Parse the GET line
                string[] lines = request.Split("\r\n");
                string getLine = lines[0];
                string[] parts = getLine.Split(' ');
                string fileName = parts.Length > 1 ? parts[1].TrimStart('/') : "";

                if (string.IsNullOrWhiteSpace(fileName) || !File.Exists(fileName))
                {
                    Console.WriteLine($"[Server] 404 - File not found: {fileName}");
                    string notFound = "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\nConnection: close\r\n\r\n";
                    byte[] notFoundBytes = Encoding.UTF8.GetBytes(notFound);
                    stream.Write(notFoundBytes, 0, notFoundBytes.Length);
                }
                else
                {
                    Console.WriteLine($"[Server] 200 - Sending file: {fileName}");
                    byte[] fileBytes = File.ReadAllBytes(fileName);
                    string header = $"HTTP/1.1 200 OK\r\nContent-Length: {fileBytes.Length}\r\nConnection: close\r\n\r\n";
                    byte[] headerBytes = Encoding.UTF8.GetBytes(header);
                    // Send headers
                    stream.Write(headerBytes, 0, headerBytes.Length);
                    // Send body
                    stream.Write(fileBytes, 0, fileBytes.Length);
                }

            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Server] Error: {ex.Message}");
            }
            finally
            {
                client.Close();
            }
        }
    }
}

