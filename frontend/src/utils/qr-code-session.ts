export interface QRCodeState {
    image: string;
    loading: boolean;
    remaining: number;
}

export const createQRCodeSession = <Params>(
    fetchImage: (params: Params) => Promise<string>,
    publish: (state: QRCodeState) => void,
    timers = {
        setInterval: (callback: () => void, delay: number) => setInterval(callback, delay),
        clearInterval: (timer: ReturnType<typeof setInterval>) => clearInterval(timer),
    },
    now = Date.now,
) => {
    let version = 0;
    let timer: ReturnType<typeof setInterval> | undefined;
    const clearTimer = () => {
        if (timer !== undefined) timers.clearInterval(timer);
        timer = undefined;
    };
    const cancel = () => {
        version++;
        clearTimer();
        publish({ image: '', loading: false, remaining: 0 });
    };
    const generate = async (params: Params) => {
        const ticket = ++version;
        clearTimer();
        publish({ image: '', loading: true, remaining: 0 });
        try {
            const image = await fetchImage(params);
            if (ticket !== version) return;
            let remaining = 60;
            const expiresAt = now() + remaining * 1000;
            publish({ image, loading: false, remaining });
            timer = timers.setInterval(() => {
                if (ticket !== version) return;
                remaining = Math.max(0, Math.ceil((expiresAt - now()) / 1000));
                if (remaining <= 0) {
                    void generate(params).catch(() => {});
                } else {
                    publish({ image, loading: false, remaining });
                }
            }, 1000);
        } catch (error) {
            if (ticket === version) {
                publish({ image: '', loading: false, remaining: 0 });
            }
            throw error;
        }
    };
    return { generate, cancel };
};
