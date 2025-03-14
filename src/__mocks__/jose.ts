export class EncryptJWT {
  private payload: any;

  constructor(payload: any) {
    this.payload = payload;
  }

  setProtectedHeader(header: { alg: string; enc: string }) {
    return this; 
  }

  async encrypt(key: Buffer) {
    return `mocked-jwe-${JSON.stringify(this.payload)}`;
  }
}
