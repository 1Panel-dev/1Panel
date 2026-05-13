import { watch, onBeforeMount, onMounted, onBeforeUnmount } from 'vue';
import { useRoute } from 'vue-router';
import { MenuStore } from '@/store';
import { DeviceType } from '@/enums/app';
import { useGlobalStore } from '@/composables/useGlobalStore';
/** 参考 Bootstrap 的响应式设计 WIDTH = 600 */
const WIDTH = 600;

/** 根据大小变化重新布局 */
export default () => {
    const route = useRoute();
    const { globalStore, isMobile } = useGlobalStore();
    const menuStore = MenuStore();
    const _isMobile = () => {
        const rect = document.body.getBoundingClientRect();
        return rect.width - 1 < WIDTH;
    };

    const _resizeHandler = () => {
        if (!document.hidden) {
            const isMobileScreen = _isMobile();
            globalStore.toggleDevice(isMobileScreen ? DeviceType.Mobile : DeviceType.Desktop);
            if (isMobileScreen) {
                menuStore.closeSidebar(true);
            }
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
        if (_isMobile()) {
            globalStore.toggleDevice(DeviceType.Mobile);
            menuStore.closeSidebar(true);
        }
    });

    onBeforeUnmount(() => {
        window.removeEventListener('resize', _resizeHandler);
    });
};
