using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Threading.Tasks;

namespace lab4_client
{
    public class ContinueWithDownloader : Downloader
    {

        class LoopState
        {
            public int ContentLength = -1;
            public bool HeadersParsed = false;
        }

        public ContinueWithDownloader(IPAddress address, int port,string hostName, string path) : base(address, port, hostName,path) { }

        override public Task Run()
        {
            var endPoint = new IPEndPoint(address, port);
            var conn = new Socket(AddressFamily.InterNetwork, SocketType.Stream, ProtocolType.Tcp);
            // Console.WriteLine($"Connecting for {path}...");

            return ConnectAsync(conn, endPoint)
                .ContinueWith(connectTask =>
                {
                    if (connectTask.IsFaulted) throw connectTask.Exception;
                    // Console.WriteLine($"[Client 2] Connected! ({path})");
                    string request = $"GET {path} HTTP/1.1\r\nHost: {hostName}\r\nConnection: close\r\n\r\n";
                    byte[] toSendBytes = Encoding.UTF8.GetBytes(request);
                    return SendAsync(conn, toSendBytes);
                })
                .Unwrap()
                .ContinueWith(sendTask =>
                {
                    if (sendTask.IsFaulted) throw sendTask.Exception; 
                    // Console.WriteLine($"Sent {sendTask.Result} bytes. ({path})");
                    var loopTcs = new TaskCompletionSource<bool>();
                    ReceiveLoop(conn, new byte[4096], loopTcs, new StringBuilder(), new LoopState()); 
                    return loopTcs.Task;
                })
                .Unwrap();
        }

        private void ReceiveLoop(Socket conn, byte[] buffer, TaskCompletionSource<bool> loopTcs, StringBuilder response, LoopState loopState)
        {
            ReceiveAsync(conn, buffer).ContinueWith(receiveTask =>
            {
                if (receiveTask.IsFaulted)
                {
                    conn.Close();
                    loopTcs.SetException(receiveTask.Exception);
                    return;
                }

                int bytesRead = receiveTask.Result;
                if (bytesRead <= 0)
                {
                    conn.Close();
                    if (!CheckBodyComplete(loopState, response, conn, loopTcs))
                    {
                        if (!loopTcs.Task.IsCompleted)
                        {
                            loopTcs.SetException(new Exception("Connection closed before full body was received."));
                        }
                    }
                    return;
                }

                response.Append(Encoding.UTF8.GetString(buffer, 0, bytesRead));

                if (!loopState.HeadersParsed)
                {
                    ParseHeaders(loopState, response.ToString());
                }

                if (loopState.HeadersParsed)
                {
                    if (CheckBodyComplete(loopState, response, conn, loopTcs))
                    {
                        return;
                    }
                }

                ReceiveLoop(conn, buffer, loopTcs, response, loopState); // Recurse
            });
        }

        private void ParseHeaders(LoopState state, string respStr)
        {
            int headerEnd = respStr.IndexOf("\r\n\r\n");
            if (headerEnd != -1)
            {
                state.HeadersParsed = true;
                string headers = respStr[..headerEnd];

                foreach (var line in headers.Split("\r\n"))
                {
                    if (line.StartsWith("Content-Length:", StringComparison.OrdinalIgnoreCase))
                    {
                        var parts = line.Split(':');
                        if (parts.Length == 2 && int.TryParse(parts[1].Trim(), out int len))
                        {
                            state.ContentLength = len;
                        }
                    }
                }
                // Console.WriteLine($"Headers parsed for {path}. Content-Length: {state.ContentLength}");
            }
        }

        private bool CheckBodyComplete(LoopState state, StringBuilder response, Socket conn, TaskCompletionSource<bool> tcs)
        {
            if (state.ContentLength == -1)
            {
                if (state.HeadersParsed)
                {
                    conn.Close();
                    tcs.SetException(new Exception($"Cannot find Content-Length for {path}."));
                    return true;
                }
                return false;
            }
            
            string respStr = response.ToString();
            int headerEnd = respStr.IndexOf("\r\n\r\n");
            if (headerEnd == -1) return false;

            int bodyStart = headerEnd + 4;
            int bodyLength = respStr.Length - bodyStart;

            if (bodyLength >= state.ContentLength)
            {
                string body = respStr.Substring(bodyStart, state.ContentLength);
                Console.WriteLine($"\n---File {path} Content---\n{body}\n");
                
                conn.Close();
                tcs.SetResult(true); // Loop is done
                return true;
            }

            return false;
        }

        private Task<bool> ConnectAsync(Socket socket, EndPoint endPoint)
        {
            var tcs = new TaskCompletionSource<bool>();
            socket.BeginConnect(endPoint, ar =>
            {
                try
                {
                    socket.EndConnect(ar);
                    tcs.SetResult(true);
                }
                catch (Exception ex) 
                { 
                    socket.Close(); 
                    tcs.SetException(ex); 
                }
            }, null);
            return tcs.Task;
        }

        private Task<int> SendAsync(Socket socket, byte[] data)
        {
            var tcs = new TaskCompletionSource<int>();
            socket.BeginSend(data, 0, data.Length, SocketFlags.None, ar =>
            {
                try
                {
                    int sent = socket.EndSend(ar);
                    tcs.SetResult(sent);
                }
                catch (Exception ex) 
                { 
                    socket.Close();
                    tcs.SetException(ex); 
                }
            }, null);
            return tcs.Task;
        }

        private Task<int> ReceiveAsync(Socket socket, byte[] buffer)
        {
            var tcs = new TaskCompletionSource<int>();
            socket.BeginReceive(buffer, 0, buffer.Length, SocketFlags.None, ar =>
            {
                try
                {
                    int received = socket.EndReceive(ar);
                    tcs.SetResult(received);
                }
                catch (Exception ex) 
                { 
                    tcs.SetException(ex); 
                }
            }, null);
            return tcs.Task;
        }
    }
}