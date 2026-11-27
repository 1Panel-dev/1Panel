export interface FileListContext {
    node: string;
    tab: string;
    path: string;
}

interface FileListResponse<Result> {
    response: Result;
    isCurrent: () => boolean;
}

export const createFileListRequest = <Params extends object, Result>(
    request: (params: Params, node: string, signal: AbortSignal) => Promise<Result>,
    getContext: () => FileListContext,
    setLoading: (loading: boolean) => void,
) => {
    let generation = 0;
    let paused = false;
    let disposed = false;
    let active:
        | {
              key: string;
              controller: AbortController;
              promise: Promise<FileListResponse<Result> | undefined>;
          }
        | undefined;

    const cancel = () => {
        generation++;
        active?.controller.abort();
        active = undefined;
        setLoading(false);
    };

    const load = (params: Params): Promise<FileListResponse<Result> | undefined> => {
        if (paused || disposed) {
            return Promise.resolve(undefined);
        }
        const snapshot = { ...params };
        const context = { ...getContext() };
        const key = JSON.stringify([context, snapshot]);
        if (active?.key === key) {
            return active.promise;
        }

        active?.controller.abort();
        const controller = new AbortController();
        const currentGeneration = ++generation;
        const isCurrent = () => {
            const current = getContext();
            return (
                generation === currentGeneration &&
                !controller.signal.aborted &&
                current.node === context.node &&
                current.tab === context.tab &&
                current.path === context.path
            );
        };
        setLoading(true);
        const promise = Promise.resolve()
            .then(() => (isCurrent() ? request(snapshot, context.node, controller.signal) : undefined))
            .then((response) => (response !== undefined && isCurrent() ? { response, isCurrent } : undefined))
            .catch((error: unknown) => {
                if (!isCurrent()) {
                    return undefined;
                }
                throw error;
            })
            .finally(() => {
                if (generation === currentGeneration) {
                    active = undefined;
                    setLoading(false);
                }
            });
        active = { key, controller, promise };
        return promise;
    };

    const pause = () => {
        paused = true;
        cancel();
    };
    const resume = () => {
        if (!disposed) paused = false;
    };
    const dispose = () => {
        disposed = true;
        pause();
    };

    return { load, cancel, pause, resume, dispose };
};
