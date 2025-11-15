using System.Net;

namespace lab4_client
{
    public abstract class Downloader
    {
        protected readonly IPAddress address;
        protected readonly int port;

        protected readonly string hostName;


        protected readonly string path;

        protected Downloader(IPAddress address, int port, string hostName, string path)
        {
            this.address = address;
            this.port = port;
            this.hostName = hostName;
            this.path = path;
        }

        public abstract Task Run();

    }
}