import { watch, onBeforeMount, onMounted, onBeforeUnmount } from 'vue';
import { useRoute } from 'vue-router';
import { MenuStore } from '@/store';
import { DeviceType } from '@/enums/app';
import { useGlobalStore } from '@/composables/useGlobalStore';
/** 与 Tailwind CSS 的 md 断点保持一致 */
const MOBILE_BREAKPOINT = 768;

/** 根据大小变化重新布局 */
export default () => {
    const route = useRoute();
    const { globalStore, isMobile } = useGlobalStore();
    const menuStore = MenuStore();
    const _isMobile = () => {
        const rect = document.body.getBoundingClientRect();
        return rect.width < MOBILE_BREAKPOINT;
    };

    const _syncLayout = () => {
        const isMobileScreen = _isMobile();
        globalStore.toggleDevice(isMobileScreen ? DeviceType.Mobile : DeviceType.Desktop);
        if (isMobileScreen) {
            menuStore.closeSidebar(true);
        }
    };

    const _resizeHandler = () => {
        if (!document.hidden) {
            _syncLayout();
        }
    };

    watch(
        () => route.name,
        () => {
            if (isMobile.value && !menuStore.isCollapse) {
                menuStore.closeSidebar(false);
            }
        },
    );

    onBeforeMount(() => {
        window.addEventListener('resize', _resizeHandler);
    });

    onMounted(() => {
        _syncLayout();
    });

    onBeforeUnmount(() => {
        window.removeEventListener('resize', _resizeHandler);
    });
};
