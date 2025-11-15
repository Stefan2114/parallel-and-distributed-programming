using System.Net.Sockets; 

using System.Net;
namespace lab4_client
{
    class Program
    {
        static async Task Main(string[] args)
        {
            string url = "http://demo.borland.com/testsite/stadyn_largepagewithimages.html";
            
            Console.WriteLine($"--- Preparing to download from {url} ---");

            var uri = new Uri(url);
            string host = uri.Host;         // "info.cern.ch"
            string path = uri.AbsolutePath;   // "/index.html"
            int port = uri.Port;           // 80

            IPAddress ipAddress;
            try
            {
                var hostEntry = await Dns.GetHostEntryAsync(host);
                ipAddress = hostEntry.AddressList.FirstOrDefault(a => a.AddressFamily == AddressFamily.InterNetwork);

                if (ipAddress == null)
                {
                    Console.WriteLine("No IPv4 address found for host.");
                    return;
                }

                Console.WriteLine($"Host {host} resolved to IP {ipAddress}");
            }
            catch (Exception ex)
            {
                Console.WriteLine($"DNS lookup failed: {ex.Message}");
                return;
            }

            int nrTasks = 10000;
            List<Task> tasks = new List<Task>(nrTasks);
            Downloader downloader = new AsyncDownloader(ipAddress, port, host, path);

            for(int i = 0; i< nrTasks; i++)
            {
                tasks.Add(downloader.Run());
            }            

            try
            {
                 await Task.WhenAll(tasks); 
            }
            catch (Exception ex)
            {
                Console.WriteLine($"One or more downloads failed: {ex.Message}");
            }

            Console.WriteLine("\n--- All downloads finished ---");
        }
    }
}