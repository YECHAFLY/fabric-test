/*
 * Copyright Xuyang Ma. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const { buildCCPOrg1, buildCCPOrg2, buildWallet, prettyJSONString} = require('../../test-application/javascript/AppUtil.js');

const myChannel = 'mychannel';
const myChaincodeName = 'auction';

async function closeAuction(ccp,wallet,user,auctionID) {
	try {

		const gateway = new Gateway();
		await gateway.connect(ccp,
			{ wallet: wallet, identity: user, discovery: { enabled: true, asLocalhost: true } });

		const network = await gateway.getNetwork(myChannel);
		const contract = network.getContract(myChaincodeName);

		let auctionString = await contract.evaluateTransaction('QueryAuction',auctionID);
		let auctionJSON = JSON.parse(auctionString);

		let statefulTxn = contract.createTransaction('CloseAuction');

		if (auctionJSON.organizations.length === 2) {
			statefulTxn.setEndorsingOrganizations(auctionJSON.organizations[0],auctionJSON.organizations[1]);
		} else {
			statefulTxn.setEndorsingOrganizations(auctionJSON.organizations[0]);
		}

		console.log('\n--> Submit Transaction: close auction');
		await statefulTxn.submit(auctionID);
		console.log('*** Result: committed');

		console.log('\n--> Evaluate Transaction: query the updated auction');
		let result = await contract.evaluateTransaction('QueryBids',auctionID);
		console.log('*** Result: Auction: ' + prettyJSONString(result.toString()));

		gateway.disconnect();
	} catch (error) {
		console.error(`******** FAILED to submit bid: ${error}`);
		process.exit(1);
	}
}

async function main() {
	try {

		if (process.argv[2] === undefined || process.argv[3] === undefined ||
            process.argv[4] === undefined) {
			console.log('Usage: node closeAuction.js org userID auctionID');
			process.exit(1);
		}

		const org = process.argv[2];
		const user = process.argv[3];
		const auctionID = process.argv[4];

		if (org === 'Org1' || org === 'org1') {
			const ccp = buildCCPOrg1();
			const walletPath = path.join(__dirname, 'wallet/org1');
			const wallet = await buildWallet(Wallets, walletPath);
			await closeAuction(ccp,wallet,user,auctionID);
		}
		else if (org === 'Org2' || org === 'org2') {
			const ccp = buildCCPOrg2();
			const walletPath = path.join(__dirname, 'wallet/org2');
			const wallet = await buildWallet(Wallets, walletPath);
			await closeAuction(ccp,wallet,user,auctionID);
		}  else {
			console.log('Usage: node closeAuction.js org userID auctionID ');
			console.log('Org must be Org1 or Org2');
		}
	} catch (error) {
		console.error(`******** FAILED to run the application: ${error}`);
		if (error.stack) {
			console.error(error.stack);
		}
		process.exit(1);
	}
}


main();
