<template>
    <DialogPro v-model="open" :title="$t('app.checkTitle')" size="large">
        <el-row>
            <el-col :span="20" :offset="2" v-if="open">
                <el-alert
                    type="error"
                    :title="$t('app.deleteHelper', [$t('menu.database')])"
                    center
                    show-icon
                    :closable="false"
                />
                <br />
                <el-descriptions :column="1" border>
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
import { Database } from '@/api/interface/database';
import { routerToName } from '@/utils/router';
import { ref } from 'vue';

interface InstallProps {
    items: Array<Database.DBResource>;
}
const installData = ref();
let open = ref(false);

const acceptParams = (props: InstallProps) => {
    installData.value = props.items;
    console.log('acceptParams', props.items);
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
