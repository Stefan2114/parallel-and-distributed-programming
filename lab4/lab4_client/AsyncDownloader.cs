using System.Net;
using System.Net.Sockets;
using System.Text;

namespace lab4_client
{
    public class AsyncDownloader : Downloader
    {

        public AsyncDownloader(IPAddress address, int port, string fileName) : base(address, port, fileName) { }


        override public async Task Run()
        {
            var endPoint = new IPEndPoint(address, port);
            var conn = new Socket(AddressFamily.InterNetwork, SocketType.Stream, ProtocolType.Tcp);

            try
            {
                Console.WriteLine($"Connecting for {fileName}...");
                await ConnectAsync(conn, endPoint);
                Console.WriteLine($"Connected! ({fileName})");

                string request = $"GET {fileName} HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n";
                byte[] toSendBytes = Encoding.UTF8.GetBytes(request);
                int sent = await SendAsync(conn, toSendBytes);

                var response = new StringBuilder();
                var buffer = new byte[4096];
                int contentLength = -1;
                bool headersParsed = false;

                while (true)
                {
                    int bytesRead = await ReceiveAsync(conn, buffer);
                    if (bytesRead <= 0)
                        break; // Connection closed

                    response.Append(Encoding.UTF8.GetString(buffer, 0, bytesRead));
                    string respStr = response.ToString();

                    if (!headersParsed)
                    {
                        ParseHeaders(respStr, ref headersParsed, ref contentLength);
                    }

                    if (CheckBodyComplete(respStr, headersParsed, contentLength))
                    {
                        break;
                    }
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Error {fileName}: {ex.Message}");
                conn.Close();
            }
        }
        

        private void ParseHeaders(string respStr, ref bool headersParsed, ref int contentLength)
        {
            int headerEnd = respStr.IndexOf("\r\n\r\n");
            if (headerEnd == -1)
                return; 

            headersParsed = true;
            string headers = respStr[..headerEnd];
            
            foreach (var line in headers.Split("\r\n"))
            {
                if (line.StartsWith("Content-Length:", StringComparison.OrdinalIgnoreCase))
                {
                    var parts = line.Split(':');
                    if (parts.Length == 2 && int.TryParse(parts[1].Trim(), out int len))
                    {
                        contentLength = len;
                    }
                }
            }
            Console.WriteLine($"Headers parsed for {fileName}. Content-Length: {contentLength}");
        }

        private bool CheckBodyComplete(string respStr, bool headersParsed, int contentLength)
        {
            if (!headersParsed)
                return false;
                
            int headerEnd = respStr.IndexOf("\r\n\r\n");
            if (headerEnd == -1) return false;

            if (contentLength == -1)
            {
                Console.WriteLine($"[Client 3] Cannot find Content-Length for {fileName}.");
                throw new Exception("No Content-Length header found.");
            }

            int bodyStart = headerEnd + 4;
            int bodyLength = respStr.Length - bodyStart;

            if (bodyLength >= contentLength)
            {
                string body = respStr.Substring(bodyStart, contentLength);
                Console.WriteLine($"\n---[Client 3] File {fileName} Content---\n{body}\n-----------------------------------\n");
                return true; 
            }

            return false; 
        }


        Task<bool> ConnectAsync(Socket socket, EndPoint endPoint)
        {
            var tcs = new TaskCompletionSource<bool>();
            socket.BeginConnect(endPoint, ar =>
            {
                try
                {
                    socket.EndConnect(ar);
                    tcs.SetResult(true);
                }
                catch (Exception ex) { tcs.SetException(ex); }
            }, null);
            return tcs.Task;
        }

        Task<int> SendAsync(Socket socket, byte[] data)
        {
            var tcs = new TaskCompletionSource<int>();
            socket.BeginSend(data, 0, data.Length, SocketFlags.None, ar =>
            {
                try
                {
                    int sent = socket.EndSend(ar);
                    tcs.SetResult(sent);
                }
                catch (Exception ex) { tcs.SetException(ex); }
            }, null);
            return tcs.Task;
        }

        Task<int> ReceiveAsync(Socket socket, byte[] buffer)
        {
            var tcs = new TaskCompletionSource<int>();
            socket.BeginReceive(buffer, 0, buffer.Length, SocketFlags.None, ar =>
            {
                try
                {
                    int received = socket.EndReceive(ar);
                    tcs.SetResult(received);
                }
                catch (Exception ex) { tcs.SetException(ex); }
            }, null);
            return tcs.Task;
        }
    }
}