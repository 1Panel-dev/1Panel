import NProgress from 'nprogress';
import 'nprogress/nprogress.css';

NProgress.configure({
    easing: 'ease',
    speed: 300,
    showSpinner: true,
    trickleSpeed: 200,
    minimum: 0.3,
});
export default NProgress;
