<template>
    <el-container>
        <el-aside>
            <Menu></Menu>
        </el-aside>
        <el-container>
            <el-header>
                <Header></Header>
            </el-header>
            <el-main>
                <section class="main-box">
                    <router-view v-slot="{ Component, route }">
                        <transition appear name="fade-transform" mode="out-in">
                            <keep-alive :include="cacheRouter">
                                <component
                                    :is="Component"
                                    :key="route.path"
                                ></component>
                            </keep-alive>
                        </transition>
                    </router-view>
                </section>
            </el-main>
            <el-footer v-if="themeConfig.footer">
                <Footer></Footer>
            </el-footer>
        </el-container>
    </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import Menu from './Menu/index.vue';
import Header from './Header/index.vue';
import Footer from './Footer/index.vue';
import cacheRouter from '@/routers/cacheRouter';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();
const themeConfig = computed(() => globalStore.themeConfig);
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
