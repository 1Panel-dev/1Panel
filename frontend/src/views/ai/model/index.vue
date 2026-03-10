<template>
    <div>
        <RouterMenu />
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import RouterMenu from '@/views/ai/model/router-menu.vue';
</script>
