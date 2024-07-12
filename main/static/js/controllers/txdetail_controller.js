import BaseController from "./base_controller";

export default class extends BaseController {
  copyTxid(e) {
    const txid = e.params.txid
    this.copyText(txid)
    $("#txCopyBtn").text('done')
    $("#txCopyBtn").addClass('c-green')
    $("#copiedSpan").removeClass('d-none')
    setInterval(function(){
      $("#txCopyBtn").text('content_copy')
      $("#txCopyBtn").removeClass('c-green')
      $("#copiedSpan").addClass('d-none')
    },1500);
  }
}
