<template>
    <el-dialog
        v-model="open"
        :title="$t('app.checkTitle')"
        width="40%"
        :close-on-click-modal="false"
        :destroy-on-close="true"
    >
        <el-row>
            <el-col :span="20" :offset="2" v-if="open">
                <el-alert
                    type="error"
                    :title="$t('app.deleteHelper', [$t('app.database')])"
                    center
                    show-icon
                    :closable="false"
                />
                <el-descriptions :column="1" border>
                    <el-descriptions-item label-class-name="label" class-name="content" min-width="60px">
                        <template #label>
                            <a href="javascript:void(0);" class="content-a" @click="toApp()">{{ $t('app.app') }}</a>
                        </template>
                        <pre>{{ installData.join('\n') }}</pre>
                    </el-descriptions-item>
                </el-descriptions>
            </el-col>
        </el-row>
    </el-dialog>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
const router = useRouter();

interface InstallProps {
    items: Array<string>;
}
const installData = ref();
let open = ref(false);

const acceptParams = (props: InstallProps) => {
    installData.value = props.items;
    installData.value.push('sdsfhjdghjdgfhsdgfhjsgfkhjsdgfhjasgdfhjasgdfjhsagdfhjsagdfashdfgaskhjdfgaskjhdf');
    open.value = true;
};

const toApp = () => {
    router.push({ name: 'AppInstalled' });
};

defineExpose({
    acceptParams,
});
</script>

<style scoped>
:deep(.label) {
    background: var(--panel-main-bg-color-10) !important;
}
:deep(.content) {
    background: var(--panel-main-bg-color-10);
}
:deep(.content-a) {
    color: var(--panel-color-primary);
}
pre {
    margin: 0;
    width: 350px;
    overflow: hidden;
    text-overflow: ellipsis;
}
:deep(.el-descriptions) {
    overflow: hidden;
    text-overflow: ellipsis;
}
</style>
