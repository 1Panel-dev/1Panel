<template>
    <!-- Lives in the layout so terminal sessions outlive the terminal route.
         Each Terminal is teleported into the page's slot while the page is mounted
         and parked off-screen (still connected, still receiving output) otherwise. -->
    <div class="terminal-host" aria-hidden="true">
        <template v-for="item in store.entries" :key="item.key + ':' + item.refresh">
            <Teleport :to="store.slots[item.key] || 'body'" :disabled="!store.slots[item.key]">
                <Terminal
                    :ref="(el: any) => store.setInstance(item.key, el)"
                    @session="(id: string) => store.setSessionId(item.key, id)"
                    @expired="store.onExpired(item.key)"
                />
            </Teleport>
        </template>
    </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import Terminal from '@/components/terminal/index.vue';
import { TerminalSessionStore } from '@/store';

const store = TerminalSessionStore();

// Sessions the agent still holds (page refresh, closed browser tab) are
// reattached right away, without waiting for the terminal page.
onMounted(() => store.restore());
</script>

<style scoped>
/* Off-screen but sized, so xterm can measure and keep rendering while parked. */
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
