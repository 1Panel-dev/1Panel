<template>
    <div>
        <el-card class="el-card">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('home.app') }}</span>
                </div>
            </template>
            <el-row :gutter="20">
                <el-col :span="12">
                    <div @click="goInstall(baseInfo.haloID)">
                        <el-card style="height: 110px" @click="goInstall(baseInfo.haloID)">
                            <el-row>
                                <el-col :span="4">
                                    <img style="width: 40px; height: 40px" src="../images/halo.jpg" alt="" />
                                </el-col>
                                <el-col :span="2"><br /></el-col>
                                <el-col :span="18">
                                    <span style="font-size: 20px; color: #214bc8">Halo</span>
                                    <div>
                                        <span class="input-help">{{ $t('home.haloInfo') }}</span>
                                    </div>
                                </el-col>
                            </el-row>
                        </el-card>
                    </div>
                </el-col>
                <el-col :span="12">
                    <div @click="goInstall(baseInfo.dateeaseID)">
                        <el-card style="height: 110px" @click="goInstall(baseInfo.dateeaseID)">
                            <el-row>
                                <el-col :span="4">
                                    <img style="width: 40px; height: 40px" src="../images/de.jpg" alt="" />
                                </el-col>
                                <el-col :span="2"><br /></el-col>
                                <el-col :span="18">
                                    <span style="font-size: 20px; color: #0070d6">Dataease</span>
                                    <div>
                                        <span class="input-help">{{ $t('home.deInfo') }}</span>
                                    </div>
                                </el-col>
                            </el-row>
                        </el-card>
                    </div>
                </el-col>
            </el-row>
            <el-row :gutter="20" style="margin-top: 20px">
                <el-col :span="12">
                    <div @click="goInstall(baseInfo.jumpserverID)">
                        <el-card style="height: 110px" @click="goInstall(baseInfo.jumpserverID)">
                            <el-row>
                                <el-col :span="4">
                                    <img style="width: 40px; height: 40px" src="../images/js.jpg" alt="" />
                                </el-col>
                                <el-col :span="2"><br /></el-col>
                                <el-col :span="18">
                                    <span style="font-size: 16px; color: #008d75">JumpServer</span>
                                    <div>
                                        <span class="input-help">{{ $t('home.jsInfo') }}</span>
                                    </div>
                                </el-col>
                            </el-row>
                        </el-card>
                    </div>
                </el-col>
                <el-col :span="12">
                    <div @click="goInstall(baseInfo.metersphereID)">
                        <el-card style="height: 110px" @click="goInstall(baseInfo.metersphereID)">
                            <el-row>
                                <el-col :span="4">
                                    <img style="width: 40px; height: 40px" src="../images/ms.jpg" alt="" />
                                </el-col>
                                <el-col :span="2"><br /></el-col>
                                <el-col :span="18">
                                    <span style="font-size: 16px; color: #723279">MeterSphere</span>
                                    <div>
                                        <span class="input-help">{{ $t('home.msInfo') }}</span>
                                    </div>
                                </el-col>
                            </el-row>
                        </el-card>
                    </div>
                </el-col>
            </el-row>
            <el-row :gutter="20" style="margin-top: 20px">
                <el-col :span="12">
                    <div @click="goInstall(baseInfo.kubeoperatorID)">
                        <el-card style="height: 110px" @click="goInstall(baseInfo.kubeoperatorID)">
                            <el-row>
                                <el-col :span="4">
                                    <img style="width: 40px; height: 40px" src="../images/ko.jpg" alt="" />
                                </el-col>
                                <el-col :span="2"><br /></el-col>
                                <el-col :span="18">
                                    <span style="font-size: 16px; color: #fc3535">KubeOperator</span>
                                    <div>
                                        <span class="input-help">{{ $t('home.koInfo') }}</span>
                                    </div>
                                </el-col>
                            </el-row>
                        </el-card>
                    </div>
                </el-col>
                <el-col :span="12">
                    <div @click="goInstall(baseInfo.kubepiID)">
                        <el-card style="height: 110px" @click="goInstall(baseInfo.kubepiID)">
                            <el-row>
                                <el-col :span="4">
                                    <img style="width: 40px; height: 40px" src="../images/kubepi.jpg" alt="" />
                                </el-col>
                                <el-col :span="2"><br /></el-col>
                                <el-col :span="18">
                                    <span style="font-size: 16px; color: #fc3535">KubePi</span>
                                    <div>
                                        <span class="input-help">{{ $t('home.kubepiInfo') }}</span>
                                    </div>
                                </el-col>
                            </el-row>
                        </el-card>
                    </div>
                </el-col>
            </el-row>
        </el-card>
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

const goInstall = (id: number) => {
    let params: { [key: string]: any } = {
        id: id,
    };
    router.push({ name: 'AppDetail', params });
};

defineExpose({
    acceptParams,
});
</script>
