<template>
    <div class="delete-file-list">
        <div class="flx-align-center mb-1" v-for="(file, index) in files" :key="index">
            <svg-icon
                className="table-icon mr-1"
                :iconName="file.isDir ? 'p-file-folder' : getIcon(file.extension || '')"
            />
            <span class="sle">{{ file.name }}</span>
        </div>
    </div>
</template>

<script setup lang="ts">
import { getIcon } from '@/utils/file';

defineProps<{ files: { name: string; isDir: boolean; extension?: string }[] }>();
</script>

<style scoped>
.delete-file-list {
    max-height: 400px;
    overflow-y: auto;
    margin-top: 15px;
}
</style>
