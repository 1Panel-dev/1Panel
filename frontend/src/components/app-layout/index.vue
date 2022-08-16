<template>
    <Layout>
        <template #menu>
            <Menu></Menu>
        </template>
        <template #header>
            <Header></Header>
        </template>
        <template #footer>
            <Footer></Footer>
        </template>
    </Layout>
</template>
<script setup lang="ts">
import Layout from '@/layout/index.vue';
import Header from './header/index.vue';
import Footer from './footer/index.vue';
import Menu from './menu/index.vue';
</script>
