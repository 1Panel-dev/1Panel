<template>
    <div v-if="showOption" class="w-full">
        <el-radio-group v-model="restart" @change="changeRestart">
            <el-radio :value="true">{{ $t('setting.restartNow') }}</el-radio>
            <el-radio :value="false">{{ $t('setting.restartLater') }}</el-radio>
        </el-radio-group>
        <span class="input-help" v-if="restart">{{ $t('xpack.node.syncProxyHelper3') }}</span>
        <span class="input-help" v-else>{{ $t('xpack.node.syncProxyHelper4') }}</span>
    </div>
</template>

<script lang="ts" setup>
import { getSettingBaseInfo } from '@/api/modules/setting';
import { getXpackProxyDocker } from '@/extensions/xpack';

const showOption = ref(false);
const restart = ref(false);

const em = defineEmits(['update:withDockerRestart']);
const props = defineProps({
    syncList: String,
});

const loadStatus = async () => {
    if (props.syncList.indexOf('SyncSystemProxy') === -1) {
        em('update:withDockerRestart', false);
        return;
    }
    await getSettingBaseInfo()
        .then((res) => {
            if (res.data.proxyType === '' || res.data.proxyType === 'close') {
                em('update:withDockerRestart', false);
                return;
            }
        })
        .catch(() => {
            em('update:withDockerRestart', false);
            return;
        });
    const res = await getXpackProxyDocker();
    if (!res) {
        em('update:withDockerRestart', false);
        return;
    }
    if (res.data.proxyDocker !== 'Enable') {
        em('update:withDockerRestart', false);
        return;
    }
    showOption.value = true;
};

const changeRestart = () => {
    em('update:withDockerRestart', restart.value);
};

onMounted(() => {
    showOption.value = false;
    loadStatus();
});
</script>
