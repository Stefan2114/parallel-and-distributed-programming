
using System.Net;
namespace lab4_client
{
    class Program
    {
        static async Task Main(string[] args)
        {
            Downloader downloader1 = new AsyncDownloader(IPAddress.Loopback, 8080, "file1");
            var t1 = downloader1.Run();
            Downloader downloader2 = new AsyncDownloader(IPAddress.Loopback, 8080, "file2");
            var t2 = downloader2.Run();
            await Task.WhenAll(t1, t2);
            Console.WriteLine("Finished");
        }
    }
}