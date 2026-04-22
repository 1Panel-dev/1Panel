export interface StepOptions {
    id?: string;
    index: number;
    beforeActive?: Function;
    beforeLeave?: Function;
    title?: string;
    description?: string;
    icon?: string;
    status?: string;
}

export interface StepperOptions {
    steps: StepOptions[];
    index: number;
    activeSet: Set<number>;
    isLoading?: boolean;
    onCancelButtonText: string;
    onFinishButtonText: string;
    prevButtonText: string;
    nextButtonText: string;
    buttonSize: string;
    footerAlign: string;
    showCancel: boolean;
    beforeActive?: Function;
    beforeLeave?: Function;
    height?: string | number;
}
