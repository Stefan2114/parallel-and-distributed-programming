using System.Net;
using System.Net.Sockets;
using System.Runtime.CompilerServices;
using System.Text;

namespace lab4_client
{
    public class CallbackDownloader : Downloader
    {


        private readonly TaskCompletionSource<bool> _tcs = new TaskCompletionSource<bool>(); 
        class DownloadState
        {
            public Socket Conn;
            public string HostName;
            public string Path;  
            public byte[] Buffer = new byte[4096];
            public StringBuilder Response = new StringBuilder();
            public bool HeadersParsed = false;
            public int ContentLength = -1;
        }

        public CallbackDownloader(IPAddress address, int port,string hostName, string path) : base(address, port,hostName, path) { }

        override public Task Run()
        {
            var endPoint = new IPEndPoint(address, port);
            var conn = new Socket(AddressFamily.InterNetwork, SocketType.Stream, ProtocolType.Tcp);

            try
            {
                var state = new DownloadState { Conn = conn, HostName = hostName, Path = path };
                // Console.WriteLine($"Connecting for {path}...");
                conn.BeginConnect(endPoint, ConnectCallback, state);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Error {path}: {ex.Message}");
            }
            return _tcs.Task;
        }
        void ConnectCallback(IAsyncResult ar)
        {
            var state = (DownloadState)ar.AsyncState!;
            try
            {
                state.Conn.EndConnect(ar);
                // Console.WriteLine($"Connected! ({state.Path})");

                string request = $"GET {state.Path} HTTP/1.1\r\nHost: {state.HostName}\r\nConnection: close\r\n\r\n";
                byte[] toSendBytes = Encoding.UTF8.GetBytes(request);

                state.Conn.BeginSend(toSendBytes, 0, toSendBytes.Length, SocketFlags.None, SendCallback, state);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Error {state.Path}: {ex.Message}");
                state.Conn.Close();   
                _tcs.SetException(ex);
            }
        }

        void SendCallback(IAsyncResult ar)
        {
            var state = (DownloadState)ar.AsyncState!;
            try
            {
                int sent = state.Conn.EndSend(ar);
                // Console.WriteLine($"Sent {sent} bytes. ({state.Path})");

                state.Conn.BeginReceive(state.Buffer, 0, state.Buffer.Length, SocketFlags.None, ReceiveCallback, state);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Error {state.Path}: {ex.Message}");
                state.Conn.Close(); 
                _tcs.SetException(ex);
            }
        }

        void ReceiveCallback(IAsyncResult ar)
        {
            var state = (DownloadState)ar.AsyncState!;
            try
            {
                int bytesRead = state.Conn.EndReceive(ar);
                if (bytesRead > 0)
                {
                    state.Response.Append(Encoding.UTF8.GetString(state.Buffer, 0, bytesRead));

                    if (!state.HeadersParsed)
                    {
                        ParseHeaders(state);
                    }

                    if (state.HeadersParsed && CheckBodyComplete(state))
                    {
                        return;
                    }

                    state.Conn.BeginReceive(state.Buffer, 0, state.Buffer.Length, SocketFlags.None, ReceiveCallback, state);
                }
                else
                {
                    if (state.HeadersParsed)
                    {
                        CheckBodyComplete(state);
                    }
                    state.Conn.Close();
                    if (!_tcs.Task.IsCompleted)
                    {
                        _tcs.SetResult(true);
                    }
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Error {state.Path}: {ex.Message}");
                state.Conn.Close();
                _tcs.SetException(ex);
            }
        }

        void ParseHeaders(DownloadState state)
        {
            string resp = state.Response.ToString();
            int headerEnd = resp.IndexOf("\r\n\r\n");
            if (headerEnd != -1)
            {
                state.HeadersParsed = true;
                string headers = resp[..headerEnd];

                var lines = headers.Split("\r\n");
                foreach (var line in lines)
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
                // Console.WriteLine($"Headers parsed for {state.Path}. Content-Length: {state.ContentLength}");
            }
        }

        bool CheckBodyComplete(DownloadState state)
        {
            string resp = state.Response.ToString();
            int headerEnd = resp.IndexOf("\r\n\r\n");
            int bodyStart = headerEnd + 4; // Skip past \r\n\r\n

            if (state.ContentLength == -1)
            {
                // This could happen if headers are parsed but no Content-Length is found
                if (state.HeadersParsed)
                {
                    Console.WriteLine($"Cannot find Content-Length for {state.Path}.");
                    state.Conn.Close();
                    _tcs.SetException(new Exception("No Content-Length header."));
                    return true;
                }
                return false;
            }

            int bodyLength = resp.Length - bodyStart;
            if (bodyLength >= state.ContentLength)
            {
                string body = resp.Substring(bodyStart, state.ContentLength);
                Console.WriteLine($"\n---File {state.Path} Content---\n{body}\n");
                state.Conn.Close();
                _tcs.SetResult(true);
                return true;
            }

            return false;
        }
    }
}