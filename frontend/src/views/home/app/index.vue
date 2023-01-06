<template>
    <div>
        <el-row style="margin-top: 30px">
            <el-col :span="4" align="right">
                <img style="width: 40px; height: 40px" src="../images/halo.jpg" alt="" />
            </el-col>
            <el-col :span="1"><br /></el-col>
            <el-col :span="19">
                <span style="font-size: 20px">Halo</span>
                <el-button
                    round
                    plain
                    type="primary"
                    @click="goInstall()"
                    style="float: right; margin-top: 15px; margin-right: 10px"
                >
                    {{ $t('app.install') }}
                </el-button>
                <div>
                    <span style="font-size: 12px">{{ $t('home.haloInfo') }}</span>
                    <el-divider />
                </div>
            </el-col>
        </el-row>
        <el-row style="margin-top: 15px">
            <el-col :span="4" align="right">
                <img style="width: 40px; height: 40px" src="../images/de.jpg" alt="" />
            </el-col>
            <el-col :span="1"><br /></el-col>
            <el-col :span="19">
                <span style="font-size: 20px">Dataease</span>
                <el-button
                    round
                    plain
                    @click="goInstall()"
                    type="primary"
                    style="float: right; margin-top: 15px; margin-right: 10px"
                >
                    {{ $t('app.install') }}
                </el-button>
                <div>
                    <span style="font-size: 12px">{{ $t('home.deInfo') }}</span>
                    <el-divider />
                </div>
            </el-col>
        </el-row>
        <el-row style="margin-top: 15px">
            <el-col :span="4" align="right">
                <img style="width: 40px; height: 40px" src="../images/js.jpg" alt="" />
            </el-col>
            <el-col :span="1"><br /></el-col>
            <el-col :span="19">
                <span style="font-size: 20px">JumpServer</span>
                <el-button
                    round
                    plain
                    @click="goInstall()"
                    type="primary"
                    style="float: right; margin-top: 15px; margin-right: 10px"
                >
                    {{ $t('app.install') }}
                </el-button>
                <div>
                    <span style="font-size: 12px">{{ $t('home.jsInfo') }}</span>
                    <el-divider />
                </div>
            </el-col>
        </el-row>
        <el-row style="margin-top: 15px">
            <el-col :span="4" align="right">
                <img style="width: 40px; height: 40px" src="../images/ms.jpg" alt="" />
            </el-col>
            <el-col :span="1"><br /></el-col>
            <el-col :span="19">
                <span style="font-size: 20px">MeterSphere</span>
                <el-button
                    round
                    plain
                    @click="goInstall()"
                    type="primary"
                    style="float: right; margin-top: 15px; margin-right: 10px"
                >
                    {{ $t('app.install') }}
                </el-button>
                <div>
                    <span style="font-size: 12px">{{ $t('home.msInfo') }}</span>
                    <el-divider />
                </div>
            </el-col>
        </el-row>
        <el-row style="margin-top: 15px">
            <el-col :span="4" align="right">
                <img style="width: 40px; height: 40px" src="../images/ko.jpg" alt="" />
            </el-col>
            <el-col :span="1"><br /></el-col>
            <el-col :span="19">
                <span style="font-size: 20px">KubeOperator</span>
                <el-button round plain type="primary" style="float: right; margin-top: 15px; margin-right: 10px">
                    {{ $t('app.install') }}
                </el-button>
                <div>
                    <span style="font-size: 12px">{{ $t('home.koInfo') }}</span>
                    <el-divider />
                </div>
            </el-col>
        </el-row>
        <el-row style="margin-top: 15px">
            <el-col :span="4" align="right">
                <img style="width: 40px; height: 40px" src="../images/kubepi.jpg" alt="" />
            </el-col>
            <el-col :span="1"><br /></el-col>
            <el-col :span="19">
                <span style="font-size: 20px">KubePi</span>
                <el-button
                    round
                    plain
                    @click="goInstall()"
                    type="primary"
                    style="float: right; margin-top: 15px; margin-right: 10px"
                >
                    {{ $t('app.install') }}
                </el-button>
                <div>
                    <span style="font-size: 12px">{{ $t('home.kubepiInfo') }}</span>
                    <el-divider />
                </div>
            </el-col>
        </el-row>
    </div>
</template>

<script lang="ts" setup>
import { Dashboard } from '@/api/interface/dashboard';
import { ref } from 'vue';
import { useRouter } from 'vue-router';
const router = useRouter();

const baseInfo = ref({
    haloID: 0,
    dateeaseID: 0,
    jumpserverID: 0,
    metersphereID: 0,
    kubeoperatorID: 0,
    kubepiID: 0,
});

const acceptParams = (base: Dashboard.BaseInfo): void => {
    baseInfo.value.haloID = base.haloID;
    baseInfo.value.dateeaseID = base.dateeaseID;
    baseInfo.value.jumpserverID = base.jumpserverID;
    baseInfo.value.metersphereID = base.metersphereID;
    baseInfo.value.kubeoperatorID = base.kubeoperatorID;
    baseInfo.value.kubepiID = base.kubepiID;
};

const goInstall = () => {
    router.push({ name: 'App' });
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.el-divider--horizontal {
    display: block;
    height: 1px;
    width: 100%;
    margin: 7px 0;
    border-top: 1px var(--el-border-color) var(--el-border-style);
}
</style>
