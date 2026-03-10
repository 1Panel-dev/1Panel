<template>
    <DialogPro v-model="open" :title="$t('app.checkTitle')" size="large">
        <el-row>
            <el-col :span="20" :offset="2" v-if="open">
                <el-alert
                    type="error"
                    :description="$t('app.deleteHelper', [$t('menu.database')])"
                    center
                    show-icon
                    :closable="false"
                />
                <br />
                <el-descriptions border :column="1">
                    <el-descriptions-item
                        v-for="(item, key) in installData"
                        :key="key"
                        label-class-name="check-label"
                        class-name="check-content"
                        min-width="60px"
                    >
                        <template #label>
                            <a href="javascript:void(0);" class="check-label-a" @click="toPage(item.type)">
                                {{ $t('menu.' + item.type) }}
                            </a>
                        </template>
                        <span class="resources">
                            {{ item.name }}
                        </span>
                    </el-descriptions-item>
                </el-descriptions>
            </el-col>
        </el-row>
    </DialogPro>
</template>
<script lang="ts" setup>
import { routerToName } from '@/utils/router';
import { ref } from 'vue';

interface InstallProps {
    items: Array<{ type: string; name: string }>;
}
const installData = ref();
let open = ref(false);

const acceptParams = (props: InstallProps) => {
    installData.value = props.items;
    open.value = true;
};

const toPage = (key: string) => {
    if (key === 'app') {
        routerToName('App');
    }
    if (key === 'website') {
        routerToName('Website');
    }
};

defineExpose({
    acceptParams,
});
</script>
