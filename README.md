# 시연 영상
Linux Foundation EMCO 프로젝트에 Migration Controller를 추가 구현하여 시연한 영상 <br/><br/>
Cluster1에 서비스를 배포할 때 리소스가 있으면 바로 배포가 진행되지만, 없을 때는 Migration Controller를 호출하여 Cluster1에서 실행되고 있는 워크로드 중 일부를 선택하여 Cluster5로 이전하고 배포를 수행하는 것을 확인할 수 있다.

이전하는 과정에서는 서비스의 중단이 없도록 하기 위해 Cluster5에서 배포가 완료된 것(Running 상태)을 확인한 후 Cluster1에서 해당 워크로드를 삭제한다.

<p align="center">
  <img src="https://github.com/dbswlgp/emco/assets/46889729/5f9c103f-3d50-4aa7-a038-680db8942695">
</p>
