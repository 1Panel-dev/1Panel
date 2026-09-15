<template>
    <div class="terminal-host" aria-hidden="true">
        <template v-for="store in stores" :key="store.$id">
            <template v-for="item in store.entries" :key="item.key + ':' + item.refresh">
                <Teleport :to="store.slots[item.key] || 'body'" :disabled="!store.slots[item.key]">
                    <Terminal
                        :ref="(el: any) => store.setInstance(item.key, el)"
                        @session="(id: string) => store.setSessionId(item.key, id)"
                        @expired="store.onExpired(item.key)"
                    />
                </Teleport>
            </template>
        </template>
    </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import Terminal from '@/components/terminal/index.vue';
import { TerminalDockSessionStore, TerminalSessionStore, TerminalStore } from '@/store';
import { getTerminalInfo } from '@/api/modules/setting';
import { useGlobalStore } from '@/composables/useGlobalStore';

const pageStore = TerminalSessionStore();
const dockStore = TerminalDockSessionStore();
const terminalStore = TerminalStore();
const stores = [pageStore, dockStore];
const { isAdmin } = useGlobalStore();

onMounted(async () => {
    try {
        const res = await getTerminalInfo();
        terminalStore.showTerminalButton = res.data.showTerminalButton !== 'Disable';
    } catch {}
    if (isAdmin.value && terminalStore.showTerminalButton) await dockStore.restore();
});
</script>

<style scoped>
.terminal-host {
    position: fixed;
    left: -10000px;
    top: 0;
    width: 1000px;
    height: 600px;
    overflow: hidden;
    visibility: hidden;
    pointer-events: none;
}
</style>
