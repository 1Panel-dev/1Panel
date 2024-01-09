<template>
    <el-page-header @back="props.back">
        <template #content>
            <span>{{ header }}</span>
            <span v-if="resource && !hideResource">
                -
                <el-tooltip v-if="resource.length > 25" :content="resource" placement="bottom">
                    <el-tag effect="dark" type="success">{{ resource.substring(0, 23) + '...' }}</el-tag>
                </el-tooltip>
                <el-tag v-else effect="dark" type="success">{{ resource }}</el-tag>
            </span>
            <el-divider v-if="slots.buttons" direction="vertical" />
            <slot v-if="slots.buttons" name="buttons"></slot>
        </template>
        <template #extra>
            <slot v-if="slots.extra" name="extra"></slot>
        </template>
    </el-page-header>
</template>

<script lang="ts" setup>
import { useSlots } from 'vue';
const slots = useSlots();

defineOptions({ name: 'DrawerHeader' });

const props = defineProps({
    header: String,
    back: Function,
    resource: String,
    hideResource: Boolean,
});
</script>
