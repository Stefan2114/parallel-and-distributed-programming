void main() throws InterruptedException {
    int[] a = {1, 3, -2};
    int[] b = {4, -1, 5};

    SharedData data = new SharedData();

    Thread multiplier = new Thread(() -> {
        for (int i = 0; i < a.length; i++) {
            data.produce(a[i] * b[i]);
        }
        data.setDone();
    });

    Thread summer = new Thread(() -> {
        int sum = 0;
        while (true) {
            Integer value = data.consume();
            if (value == null) break;
            sum += value;
        }
        System.out.println("Dot product: " + sum);
    });

    multiplier.start();
    summer.start();

    multiplier.join();
    summer.join();
}
