db = db.getSiblingDB('backend_challenge');

db.createCollection('users');
db.createCollection('lottery_tickets');

db.users.createIndex({ email: 1 }, { unique: true });
db.lottery_tickets.createIndex({ number: 1, set: 1 }, { unique: true, name: 'uniq_number_set' });
db.lottery_tickets.createIndex({ reservedUntil: 1 }, { name: 'reserved_until' });
db.lottery_tickets.createIndex({ d1: 1 }, { name: 'digit1' });
db.lottery_tickets.createIndex({ d2: 1 }, { name: 'digit2' });
db.lottery_tickets.createIndex({ d3: 1 }, { name: 'digit3' });
db.lottery_tickets.createIndex({ d4: 1 }, { name: 'digit4' });
db.lottery_tickets.createIndex({ d5: 1 }, { name: 'digit5' });
db.lottery_tickets.createIndex({ d6: 1 }, { name: 'digit6' });
db.lottery_tickets.createIndex({ rand: 1 }, { name: 'rand' });
db.lottery_tickets.createIndex({ d1: 1, d2: 1, d3: 1, d4: 1, d5: 1, d6: 1, rand: 1 }, { name: 'digits_all' });


// Seed 2 million lottery tickets. !! JUST FOR RUN DOCKER FIRST TIME !!
// if want to seed more tickets, change the values of sets
// if want to re-reed tickets, remove docker volume first
const sets = 2;
const totalTicketsPerSet = 1_000_000;
const batchSize = 10_000;

function numberToDigits(number) {
    const d = new Array(6);
    for (let i = 5; i >= 0; i--) {
        d[i] = number % 10;
        number = Math.floor(number / 10);
    }
    return d;
}

for (let set = 1; set <= sets; set++) {
    let batch = [];
    for (let i = 0; i < totalTicketsPerSet; i++) {
        const digits = numberToDigits(i);
        const numberStr = digits.join('');
        batch.push({
            number: numberStr,
            set: set,
            d1: digits[0],
            d2: digits[1],
            d3: digits[2],
            d4: digits[3],
            d5: digits[4],
            d6: digits[5],
            rand: Math.random(),
        });
        if (batch.length === batchSize) {
            try {
                db.lottery_tickets.insertMany(batch);
                print(`Inserted batch ${Math.floor(i / batchSize) + 1} of set ${set}`);
            } catch (error) {
                print(`Error inserting batch ${Math.floor(i / batchSize) + 1} of set ${set}: ${error}`);
            }
            batch = [];
        }
    }
    // Insert any remaining tickets in the last partial batch for this set
    if (batch.length > 0) {
        try {
            db.lottery_tickets.insertMany(batch);
            print(`Inserted final batch of set ${set}`);
        } catch (error) {
            print(`Error inserting final batch of set ${set}: ${error}`);
        }
    }
}