<template>
    <router-view v-slot="{ Component, route }" :key="key">
        <transition appear name="fade-transform" mode="out-in">
            <keep-alive :include="cacheRouter">
                <component :is="Component" :key="route.path"></component>
            </keep-alive>
        </transition>
    </router-view>
</template>

<script setup lang="ts">
import cacheRouter from '@/routers/cache-router';
import { useRoute } from 'vue-router';
import { computed } from 'vue';

const key = computed(() => {
    return useRoute().path + Math.random();
});
</script>
