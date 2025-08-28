import { nextTick, ref, watch } from 'vue';

export function useSearchable(paths) {
    const searchableStatus = ref(false);
    const searchablePath = ref('');
    const searchableInputRef = ref();

    watch(searchableStatus, (val) => {
        if (val) {
            searchablePath.value = paths.value.at(-1)?.url;
            nextTick(() => {
                searchableInputRef.value?.focus();
            });
        }
    });
    const searchableInputBlur = () => {
        searchableStatus.value = false;
    };

    return {
        searchableStatus,
        searchablePath,
        searchableInputRef,
        searchableInputBlur,
    };
}

export function useSearchableForSelect(paths) {
    const searchableStatus = ref(false);
    const searchablePath = ref('');
    const searchableInputRef = ref();

    watch(searchableStatus, (val) => {
        if (val) {
            if (paths.value.length === 0) {
                searchablePath.value = '/';
            } else {
                searchablePath.value = '/' + paths.value.join('/');
            }
            nextTick(() => {
                searchableInputRef.value?.focus();
            });
        }
    });
    const searchableInputBlur = () => {
        searchableStatus.value = false;
    };

    return {
        searchableStatus,
        searchablePath,
        searchableInputRef,
        searchableInputBlur,
    };
}

export function useMultipleSearchable(paths) {
    const searchableStatus = ref(false);
    const searchablePath = ref('');
    const searchableInputRefs = ref<Record<string, HTMLInputElement | null>>({});

    const setSearchableInputRef = (id: string, el: HTMLInputElement | null) => {
        if (el) {
            searchableInputRefs.value[id] = el;
        } else {
            delete searchableInputRefs.value[id];
        }
    };

    watch(searchableStatus, (val) => {
        if (val) {
            searchablePath.value = paths.value.at(-1)?.url || '';
            nextTick(() => {
                const keys = Object.keys(searchableInputRefs.value);
                if (keys.length > 0) {
                    const lastKey = keys[keys.length - 1];
                    searchableInputRefs.value[lastKey]?.focus();
                }
            });
        }
    });

    const searchableInputBlur = () => {
        searchableStatus.value = false;
    };

    return {
        searchableStatus,
        searchablePath,
        searchableInputRefs,
        setSearchableInputRef,
        searchableInputBlur,
    };
}
