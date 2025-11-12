using System.Net;

namespace lab4_client
{
    public abstract class Downloader
    {
        protected readonly IPAddress address;
        protected readonly int port;

        protected readonly string fileName;

        protected Downloader(IPAddress address, int port, string fileName)
        {
            this.address = address;
            this.port = port;
            this.fileName = fileName;
        }

        public abstract Task Run();

    }
}