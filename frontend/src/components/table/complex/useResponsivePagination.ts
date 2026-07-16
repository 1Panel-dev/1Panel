import { computed, onBeforeUnmount, onMounted, ref, type Ref } from 'vue';

interface PaginationConfig {
    small?: boolean;
}

export const useResponsivePagination = (
    paginationConfig: () => PaginationConfig | undefined,
    isMobile: Ref<boolean>,
) => {
    const paginationRef = ref<HTMLElement | null>(null);
    const paginationWidth = ref(0);
    let resizeObserver: ResizeObserver | null = null;

    const updateWidth = () => {
        paginationWidth.value = paginationRef.value?.clientWidth || 0;
    };

    const responsivePaginationLayout = computed(() => {
        if (isMobile.value || paginationConfig()?.small || paginationWidth.value < 520) {
            return 'total, prev, pager, next';
        }
        return 'total, sizes, prev, pager, next, jumper';
    });

    const responsivePagerCount = computed(() =>
        isMobile.value || paginationConfig()?.small || paginationWidth.value < 720 ? 5 : 7,
    );

    onMounted(() => {
        updateWidth();
        if (paginationRef.value) {
            resizeObserver = new ResizeObserver(updateWidth);
            resizeObserver.observe(paginationRef.value);
        }
    });
    onBeforeUnmount(() => resizeObserver?.disconnect());

    return { paginationRef, responsivePaginationLayout, responsivePagerCount };
};
