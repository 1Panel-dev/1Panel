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
